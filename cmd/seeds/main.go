package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/seeds"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	// Init Context and Config
	ctx := context.Background()
	cfg := config.MustLoad()

	// Init DB
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	// Init Location (Timezone)
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}

	// Init SQLC
	queries := sqlc.New(conn)

	seeder := seeds.New(queries, loc)
	if err := seeder.Seed(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✨ test hikes inserted successfully")
}
