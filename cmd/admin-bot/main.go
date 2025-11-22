package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/admin"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func main() {
	// TODO: init Logger

	// Init Context and Config
	ctx := context.Background()
	cfg := config.MustLoad()

	// Init Location (Timezone)
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}

	// Init TelegramBotAPI
	bot, err := tgbot.NewBotAPI(cfg.AdminBotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = false

	// Init DB
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	// Init SQLC
	queries := sqlc.New(conn)

	// Init router
	r := admin.NewRouter(loc, bot, queries, cfg.AdminChatID)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.Printf("route error: %v", err)
		}
	}
}
