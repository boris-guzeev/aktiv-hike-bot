package handler

import (
	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	adminService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/admin/service"
	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/service"
	hikeService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/user/service"
)

type Handler struct {
	bot            *tgbot.BotAPI
	cfg            config.ClientBot
	userService    userService.Service
	adminService   adminService.Service
	hikeService    hikeService.Service
	bookingService bookingService.Service
}

func New(
	b *tgbot.BotAPI,
	c config.ClientBot,
	uS userService.Service,
	aS adminService.Service,
	hS hikeService.Service,
	bS bookingService.Service,
) *Handler {
	return &Handler{
		bot:            b,
		cfg:            c,
		userService:    uS,
		adminService:   aS,
		hikeService:    hS,
		bookingService: bS,
	}
}
