package handler

import (
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
)

type HikeHandler struct {
	bot         *tgbot.BotAPI
	fsm         *fsm.FSM
	service     service.Service
	storageRoot string
	loc         *time.Location
}

func New(b *tgbot.BotAPI, s service.Service, sroot string, l *time.Location) *HikeHandler {
	return &HikeHandler{
		bot:         b,
		fsm:         fsm.NewFSM(),
		service:     s,
		storageRoot: sroot,
		loc:         l,
	}
}
