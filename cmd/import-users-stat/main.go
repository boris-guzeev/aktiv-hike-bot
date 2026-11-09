package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type databaseUser struct {
	ID       int32
	Username string
}

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

	matchedStats := make([]userStat, 0, len(parsed.Stats))
	matchedUsers := make([]databaseUser, 0, len(parsed.Stats))
	for _, stat := range parsed.Stats {
		matches := users[stat.Username]
		switch len(matches) {
		case 0:
			parsed.Issues = append(parsed.Issues, importIssue{Row: stat.Row, Message: fmt.Sprintf("user @%s not found", stat.Username)})
		case 1:
			matchedStats = append(matchedStats, stat)
			matchedUsers = append(matchedUsers, matches[0])
		default:
			parsed.Issues = append(parsed.Issues, importIssue{Row: stat.Row, Message: fmt.Sprintf("multiple database users match @%s", stat.Username)})
		}
	}

	if dryRun {
		printReport(parsed.RowsRead, len(matchedStats), 0, true, parsed.Issues)
		return nil
	}

	updated := 0
	for i, stat := range matchedStats {
		result, err := tx.Exec(ctx, `
UPDATE telegram_users
SET initial_distance_km = $1::numeric,
    initial_elevation_gain_m = $2,
    initial_hikes_count = $3
WHERE id = $4
`, stat.DistanceKm, stat.ElevationGainM, stat.HikesCount, matchedUsers[i].ID)
		if err != nil {
			return fmt.Errorf("row %d: update @%s: %w", stat.Row, stat.Username, err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("row %d: update @%s affected %d rows", stat.Row, stat.Username, result.RowsAffected())
		}
		updated++
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	printReport(parsed.RowsRead, len(matchedStats), updated, false, parsed.Issues)
	return nil
}

func printReport(rowsRead, matched, updated int, dryRun bool, issues []importIssue) {
	if dryRun {
		fmt.Println("Import report (dry run; no data changed):")
	} else {
		fmt.Println("Import report:")
	}
	fmt.Printf("  Rows read: %d\n", rowsRead)
	if dryRun {
		fmt.Printf("  Rows ready to update: %d\n", matched)
	} else {
		fmt.Printf("  Rows updated: %d\n", updated)
	}
	fmt.Printf("  Rows skipped: %d\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("  - row %d: %s\n", issue.Row, issue.Message)
	}
}

func loadUsers(ctx context.Context, tx pgx.Tx) (map[string][]databaseUser, error) {
	rows, err := tx.Query(ctx, `
SELECT id, tg_username
FROM telegram_users
WHERE tg_username IS NOT NULL
`)
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
		if strings.TrimSpace(username) == "" {
			continue
		}
		users[username] = append(users[username], user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}
	return users, nil
}
