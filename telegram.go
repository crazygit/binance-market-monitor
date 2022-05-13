package main

import (
	"github.com/crazygit/BinanceMarketMonitor/helper"
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
		log.WithField("name", bot.Self.UserName).Debug("Init Bot")
	}
}

func PostMessageToTgChannel(username, text string) error {
	msg := tgbotapi.NewMessageToChannel(username, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	response, err := tgBot.Send(msg)
	// todo: 查看返回的response格式是怎么样的,该如何处理. 非200的情况下，是有也是返回response而不是error
	log.WithField("response", response)
	return err

}
