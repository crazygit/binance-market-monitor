package main

import (
	"github.com/crazygit/binance-market-monitor/helper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var tgBot *tgbotapi.BotAPI

func init() {
	token := helper.GetRequiredStringEnv("TELEGRAM_API_TOKEN")
	if bot, err := tgbotapi.NewBotAPI(token); err != nil {
		panic(err)
	} else {
		bot.Debug = !helper.IsProductionEnvironment()
		tgBot = bot
		log.WithField("name", bot.Self.UserName).Info("Init telegram Bot")
	}
}

func PostMessageToTgChannel(username, text string) error {
	msg := tgbotapi.NewMessageToChannel(username, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if helper.IsProductionEnvironment() {
		response, err := tgBot.Send(msg)
		log.WithField("response", response).Info("Post returned message")
		return err
	} else {
		log.Info(msg)
		return nil
	}
}
