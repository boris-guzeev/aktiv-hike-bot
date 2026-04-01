package main

import (
	"context"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"

	hikeHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/handler"
	hikeRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/repository"
	hikeService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"

	userRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/repository"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"

	bookingHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/handler"
	bookingRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/repository"
	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/service"
)

func main() {
	// TODO: вынести логгер отдельно
	var log = logger.InitLogger()

	// Init Context and Config
	ctx := context.Background()
	cfg := config.MustLoadAdminBot()

	// Init Location (Timezone)
	loc, err := time.LoadLocation(cfg.Timezone)
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
	hikeHnd := hikeHandler.New(bot, hikeSvc, cfg.StorageRoot, loc)

	// --- User --- /
	userRepo := userRepository.New(queries)
	userSvc := userService.New(userRepo)

	// --- Booking --- /
	bookingRepo := bookingRepository.New(queries)
	bookingSvc := bookingService.New(bookingRepo)
	bookingHnd := bookingHandler.New(bot, userSvc, bookingSvc)

	// Init router
	r := adminbot.NewRouter(bot, cfg.AdminChatID, hikeHnd, bookingHnd)

	u := tgbot.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.StructuredError("admin-bot error", err)
		}
	}
}
