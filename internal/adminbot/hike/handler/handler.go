package handler

import (
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HikeHandler struct {
	bot     *tgbot.BotAPI
	fsm     *fsm.FSM
	service service.Service
	loc     *time.Location
}

func New(b *tgbot.BotAPI, s service.Service, l *time.Location) *HikeHandler {
	return &HikeHandler{
		bot:     b,
		fsm:     fsm.NewFSM(),
		service: s,
		loc:     l,
	}
}
