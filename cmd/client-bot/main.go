package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/config"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	notificationRepository "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/notification/repository"
	notificationWorker "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/notification/worker"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/fsm"
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
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

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
	fsm := fsm.New()
	hikeRep := hikeRepository.New(queries)
	hikeSrv := hikeService.New(hikeRep)
	hikeHnd := hikeHandler.New(bot, fsm, cfg, hikeSrv)

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
	router := clientbot.NewRouter(bot, cfg, hikeHnd, bookHnd)

	// Init Notification Worker
	workerRepo := notificationRepository.New(log, queries)
	worker := notificationWorker.New(log, bot, workerRepo)
	go worker.Run(ctx)

	// Bot updates
	updateConfig := tgbot.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)
	defer bot.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			log.Info("client bot shutting down")
			return

		case update, ok := <-updates:
			if !ok {
				log.Info("telegram updates channel closed")
				return
			}
			if err := router.Route(ctx, update); err != nil {
				log.StructuredError("bot route error", err)
			}
		}
	}
}
