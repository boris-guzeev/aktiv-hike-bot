package handler

import (
	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbot.BotAPI
	cfg     config.ClientBot
	service service.Service
}

func New(b *tgbot.BotAPI, c config.ClientBot, s service.Service) *Handler {
	return &Handler{
		bot:     b,
		cfg:     c,
		service: s,
	}
}
