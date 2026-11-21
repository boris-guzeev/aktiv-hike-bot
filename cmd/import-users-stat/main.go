package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type databaseUser struct {
	ID       int32
	Username string
}
type achievement struct {
	ID       int16
	Name     string
	HasValue bool
}

var seasonsPattern = regexp.MustCompile(`(?i)^(?:(\d+|[IVXLCDM]+)\s+)?SEASONS?$`)

func main() {
	filePath := flag.String("file", "users-initial-stat.csv", "path to the CSV file")
	dryRun := flag.Bool("dry-run", false, "validate and match users without updating the database")
	flag.Parse()
	if err := run(context.Background(), *filePath, *dryRun); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, filePath string, dryRun bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open CSV: %w", err)
	}
	defer file.Close()
	parsed, err := parseUserStats(file)
	if err != nil {
		return err
	}

	_ = godotenv.Load(".env")
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return fmt.Errorf("missing env DB_DSN")
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	users, err := loadUsers(ctx, tx)
	if err != nil {
		return err
	}
	achievements, err := loadAchievements(ctx, tx)
	if err != nil {
		return err
	}

	ready, updated, created := 0, 0, 0
	warnings := make([]importIssue, 0)
	for _, stat := range parsed.Stats {
		matches := users[stat.Username]
		if len(matches) > 1 {
			parsed.Issues = append(parsed.Issues, importIssue{Row: stat.Row, Message: fmt.Sprintf("multiple database users match @%s", stat.Username)})
			continue
		}
		ready++
		if dryRun {
			_, unknown := resolveAchievements(stat.Achievements, achievements)
			for _, name := range unknown {
				warnings = append(warnings, importIssue{Row: stat.Row, Message: fmt.Sprintf("unknown achievement %q", name)})
			}
			continue
		}

		var userID int32
		if len(matches) == 0 {
			err = tx.QueryRow(ctx, `
INSERT INTO telegram_users (tg_user_id, tg_username)
VALUES (-nextval(pg_get_serial_sequence('telegram_users', 'id')), $1)
RETURNING id`, stat.Username).Scan(&userID)
			if err != nil {
				return fmt.Errorf("row %d: create @%s: %w", stat.Row, stat.Username, err)
			}
			created++
		} else {
			userID = matches[0].ID
		}

		result, err := tx.Exec(ctx, `
UPDATE telegram_users
SET initial_distance_km = $1::numeric,
    initial_elevation_gain_m = $2,
    initial_points = $3,
    initial_hikes_count = $4
WHERE id = $5`, stat.DistanceKm, stat.ElevationGainM, stat.Points, stat.Events, userID)
		if err != nil {
			return fmt.Errorf("row %d: update @%s: %w", stat.Row, stat.Username, err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("row %d: update @%s affected %d rows", stat.Row, stat.Username, result.RowsAffected())
		}

		resolved, unknown := resolveAchievements(stat.Achievements, achievements)
		for _, name := range unknown {
			warnings = append(warnings, importIssue{Row: stat.Row, Message: fmt.Sprintf("unknown achievement %q", name)})
		}
		for achievementID, value := range resolved {
			if value > 0 {
				_, err = tx.Exec(ctx, `
INSERT INTO telegram_users_to_achievements (telegram_user_id, achievement_id, value)
VALUES ($1, $2, $3)
ON CONFLICT (telegram_user_id, achievement_id) DO UPDATE
SET value = GREATEST(telegram_users_to_achievements.value, EXCLUDED.value)`, userID, achievementID, value)
			} else {
				_, err = tx.Exec(ctx, `INSERT INTO telegram_users_to_achievements (telegram_user_id, achievement_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, achievementID)
			}
			if err != nil {
				return fmt.Errorf("row %d: add achievement for @%s: %w", stat.Row, stat.Username, err)
			}
		}
		updated++
	}

	if dryRun {
		printReport(parsed.RowsRead, ready, 0, 0, true, parsed.Issues, warnings)
		return nil
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	printReport(parsed.RowsRead, ready, updated, created, false, parsed.Issues, warnings)
	return nil
}

func printReport(rowsRead, ready, updated, created int, dryRun bool, issues, warnings []importIssue) {
	if dryRun {
		fmt.Println("Import report (dry run; no data changed):")
	} else {
		fmt.Println("Import report:")
	}
	fmt.Printf("  Rows read: %d\n", rowsRead)
	if dryRun {
		fmt.Printf("  Rows ready to import: %d\n", ready)
	} else {
		fmt.Printf("  Rows updated: %d\n  Users created: %d\n", updated, created)
	}
	fmt.Printf("  Rows skipped: %d\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("  - row %d: %s\n", issue.Row, issue.Message)
	}
	if len(warnings) > 0 {
		fmt.Printf("  Warnings: %d\n", len(warnings))
		for _, warning := range warnings {
			fmt.Printf("  - row %d: %s\n", warning.Row, warning.Message)
		}
	}
}

func loadUsers(ctx context.Context, tx pgx.Tx) (map[string][]databaseUser, error) {
	rows, err := tx.Query(ctx, `SELECT id, tg_username FROM telegram_users WHERE tg_username IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("load users: %w", err)
	}
	defer rows.Close()
	users := make(map[string][]databaseUser)
	for rows.Next() {
		var user databaseUser
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		username := normalizeUsername(user.Username)
		if username != "" {
			users[username] = append(users[username], user)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}
	return users, nil
}

func loadAchievements(ctx context.Context, tx pgx.Tx) (map[string]achievement, error) {
	rows, err := tx.Query(ctx, `SELECT id, name, has_value FROM achievements`)
	if err != nil {
		return nil, fmt.Errorf("load achievements: %w", err)
	}
	defer rows.Close()
	result := make(map[string]achievement)
	for rows.Next() {
		var a achievement
		if err := rows.Scan(&a.ID, &a.Name, &a.HasValue); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		result[normalizeAchievementName(a.Name)] = a
	}
	return result, rows.Err()
}

func resolveAchievements(names []string, known map[string]achievement) (map[int16]int32, []string) {
	result := make(map[int16]int32)
	var unknown []string
	for _, raw := range names {
		if match := seasonsPattern.FindStringSubmatch(strings.TrimSpace(raw)); match != nil {
			a, ok := known["SEASONS"]
			value, valid := parseSeasonValue(match[1])
			if !ok || !valid {
				unknown = append(unknown, raw)
				continue
			}
			if value > result[a.ID] {
				result[a.ID] = value
			}
			continue
		}
		a, ok := known[normalizeAchievementName(raw)]
		if !ok || a.HasValue {
			unknown = append(unknown, raw)
			continue
		}
		result[a.ID] = 0
	}
	return result, unknown
}

func normalizeAchievementName(value string) string {
	return strings.ToUpper(strings.Join(strings.Fields(value), " "))
}

func parseSeasonValue(value string) (int32, bool) {
	if value == "" {
		return 0, false
	}
	if n, err := strconv.ParseInt(value, 10, 32); err == nil && n > 0 {
		return int32(n), true
	}
	roman := map[byte]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	value = strings.ToUpper(value)
	total, previous := 0, 0
	for i := len(value) - 1; i >= 0; i-- {
		n, ok := roman[value[i]]
		if !ok {
			return 0, false
		}
		if n < previous {
			total -= n
		} else {
			total += n
			previous = n
		}
	}
	return int32(total), total > 0
}
