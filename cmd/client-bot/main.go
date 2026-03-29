package main

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	hikeHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/handler"
	hikeRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/repository"
	hikeService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"

	userRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/user/repository"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/user/service"

	adminRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/admin/repository"
	adminService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/admin/service"

	bookingHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/handler"
	bookingRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/repository"
	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/service"
)

func main() {
	// Init Logger
	log := logger.InitLogger()

	// Init Context
	ctx := context.Background()

	// Get Config
	cfg := config.MustLoadClientBot()

	// Init TelegramBotAPI
	bot, err := tgbot.NewBotAPI(cfg.ClientBotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = false

	// Init DB
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Init SQLC
	queries := sqlc.New(pool)

	// Init Application Dependencies
	// --- Hike --- /
	hikeRep := hikeRepository.New(queries)
	hikeSrv := hikeService.New(hikeRep)
	hikeHnd := hikeHandler.New(bot, cfg, hikeSrv)

	// --- User --- /
	userRepo := userRepository.New(queries)
	userSrv := userService.New(userRepo)

	// --- Admin --- /
	adminRepo := adminRepository.New(queries)
	adminSrv := adminService.New(adminRepo)

	// --- Booking --- /
	bookRepo := bookingRepository.New(queries)
	bookSrv := bookingService.New(bookRepo)
	bookHnd := bookingHandler.New(bot, cfg, userSrv, adminSrv, hikeSrv, bookSrv)

	// Init Router
	r := clientbot.NewRouter(bot, cfg, hikeHnd, bookHnd)

	// Bot updates
	u := tgbot.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)
	defer bot.StopReceivingUpdates()

	for upd := range updates {
		if err := r.Route(ctx, upd); err != nil {
			log.StructuredError("bot error", err)
		}
	}
}
