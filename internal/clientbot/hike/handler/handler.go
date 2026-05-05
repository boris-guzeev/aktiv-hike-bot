package handler

import (
	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/config"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbot.BotAPI
	fsm     *fsm.FSM
	cfg     config.ClientBot
	service service.Service
}

func New(b *tgbot.BotAPI, fsm *fsm.FSM, c config.ClientBot, s service.Service) *Handler {
	return &Handler{
		bot:     b,
		fsm:     fsm,
		cfg:     c,
		service: s,
	}
}
