package main

import (
	"context"
	"os"
	"strconv"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	// TODO: init Logger
	log := logrus.New()
	// TODO: init Config
	// TODO: собрать все переменные в конфиг
	adminChatIdStr := os.Getenv("ADMIN_CHAT_ID")
	adminChatId, err := strconv.ParseInt(adminChatIdStr, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	// Init Context
	ctx := context.Background()

	// Get Config
	cfg := config.MustLoadClientBot()

	// Init TelegramBotAPI
	botToken := os.Getenv("CLIENT_BOT_TOKEN")
	clientBot, err := tgbot.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}
	clientBot.Debug = false

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
	r := clientbot.NewRouter(log, clientBot, queries, adminChatId)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := clientBot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.Printf("route error: %v", err)
		}
	}
}