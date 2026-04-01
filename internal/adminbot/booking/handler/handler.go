package booking

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/service"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"
)

type BookingHandler struct {
	bot            *tgbot.BotAPI
	userService    userService.Service
	bookingService bookingService.Service
}

func New(b *tgbot.BotAPI, uS userService.Service, bS bookingService.Service) *BookingHandler {
	return &BookingHandler{
		bot:            b,
		userService:    uS,
		bookingService: bS,
	}
}
