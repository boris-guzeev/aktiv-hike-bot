package commands

import (
	"context"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Setup(_ context.Context, bot *tgbot.BotAPI, adminChatID int64) error {
	_, err := bot.Request(tgbot.NewSetMyCommandsWithScope(
		tgbot.NewBotCommandScopeChat(adminChatID),
		[]tgbot.BotCommand{
			{Command: "help", Description: "Помощь"},
			{Command: "hikes", Description: "Список хайков"},
			{Command: "newhike", Description: "Создать хайк"},
		}...,
	))
	if err != nil {
		return err
	}
	return err
}

func DeleteAllPrivateCommands(bot *tgbot.BotAPI) error {
    _, err := bot.Request(tgbot.NewDeleteMyCommandsWithScope(
        tgbot.NewBotCommandScopeAllPrivateChats(),
    ))
    return err
}