package main

import (
	"context"
	"os"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot"
	hikeHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/handler"
	hikeRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/repository"
	hikeService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	// TODO: вынести логгер отдельно
	var log = logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)                     //default
	log.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout

	// Init Context and Config
	ctx := context.Background()
	cfg := config.MustLoadAdminBot()

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

	// Init application dependencies
	// --- Hike --- /
	hikeRep := hikeRepository.New(queries)
	hikeSvc := hikeService.New(hikeRep)
	hikeHnd := hikeHandler.New(bot, hikeSvc, loc)

	// Init router
	r := adminbot.NewRouter(bot, cfg.AdminChatID, hikeHnd)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.Printf("route error: %v", err)
		}
	}
}
