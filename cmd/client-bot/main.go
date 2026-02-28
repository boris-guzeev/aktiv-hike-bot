package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func main() {
	// TODO: init Logger

	// Init Context
	ctx := context.Background()

	// Init Location (Timezone)
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}

	// Init TelegramBotAPI
	botToken := os.Getenv("CLIENT_BOT_TOKEN")
	bot, err := tgbot.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = false

	// Init DB
	dbDsn := os.Getenv("DB_DSN")
	conn, err := pgx.Connect(ctx, dbDsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	// Init SQLC
	queries := sqlc.New(conn)

	// Init router
	r := client.NewRouter(loc, bot, queries)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.Printf("route error: %v", err)
		}
	}
}