package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/commands"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/router"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func main() {
	// Init Context and Config
	ctx := context.Background()
	cfg := config.MustLoad()

	// Init Location (Timezone)
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}

	// Init TelegramBotAPI
	bot, err := tgbot.NewBotAPI(cfg.BotToken)
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

	//Setup commands
	commands.DeleteAllPrivateCommands(bot)
	err = commands.Setup(ctx, bot, cfg.AdminChatID)
	if err != nil {
		log.Fatalf("set commands: %v", err)
	}

	// Init router
	r := router.New(loc, bot, queries, cfg.AdminChatID)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Handle(ctx, upd); err != nil {
			log.Printf("handle error: %v", err)
		}
	}
}