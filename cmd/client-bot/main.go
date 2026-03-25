package main

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func main() {
	// TODO: init Logger
	log := logger.InitLogger()

	// Init Context
	ctx := context.Background()

	// Get Config
	cfg := config.MustLoadClientBot()

	// Init TelegramBotAPI
	clientBot, err := tgbot.NewBotAPI(cfg.ClientBotToken)
	if err != nil {
		log.Fatal(err)
	}
	clientBot.Debug = false

	// Init DB
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	// Init SQLC
	queries := sqlc.New(conn)

	// Init router
	r := clientbot.NewRouter(log, clientBot, queries, cfg)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := clientBot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.StructuredError("bot error", err)
		}
	}
}
