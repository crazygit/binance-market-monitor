package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/crazygit/BinanceMarketMonitor/helper"
	l "github.com/crazygit/BinanceMarketMonitor/helper/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"math"
	"strconv"
	"strings"
	"time"
)

var log = l.GetLog()

var (
	baseAssets = []string{"btc", "eth", "bnb"}
	quoteAsset = "usdt"
)

var lastAlert = map[string]ExtendWsMarketStatEvent{}

type ExtendWsMarketStatEvent struct {
	*binance.WsMarketStatEvent
	PriceChangeFloat float64
}

func escapeTextToMarkdownV2(text string) string {
	return tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, text)
}

func (e ExtendWsMarketStatEvent) AlertText() string {
	return fmt.Sprintf(`
*Symbol*: %s
*PriceChangePercent*: %s
*LastPrice*: %s
`, escapeTextToMarkdownV2(prettySymbol(e.Symbol)),
		escapeTextToMarkdownV2(prettyFloatString(e.PriceChangePercent)+"%"),
		escapeTextToMarkdownV2("$"+prettyFloatString(e.LastPrice)),
	)
}

func prettyFloatString(value string) string {
	if p, err := strconv.ParseFloat(value, 64); err != nil {
		return value
	} else {
		return fmt.Sprintf("%.2f", p)
	}
}

func prettySymbol(symbol string) string {
	var replacer *strings.Replacer
	replacer = strings.NewReplacer(strings.ToUpper(quoteAsset), fmt.Sprintf("/%s", strings.ToUpper(quoteAsset)))
	return replacer.Replace(symbol)
}

func isNeedAlert(newEvent ExtendWsMarketStatEvent) bool {
	if oldEvent, ok := lastAlert[newEvent.Symbol]; ok {
		return math.Abs(oldEvent.PriceChangeFloat-newEvent.PriceChangeFloat) >= 5
	} else {
		lastAlert[newEvent.Symbol] = newEvent
	}
	return false
}

func eventHandler(event *binance.WsMarketStatEvent) {
	priceChangeFloat, _ := strconv.ParseFloat(event.PriceChangePercent, 64)
	newEvent := ExtendWsMarketStatEvent{WsMarketStatEvent: event, PriceChangeFloat: priceChangeFloat}
	log.WithFields(logrus.Fields{
		"Symbol":             newEvent.Symbol,
		"PriceChange":        prettyFloatString(newEvent.LastPrice),
		"PriceChangePercent": newEvent.PriceChangePercent,
		"LastPrice":          prettyFloatString(newEvent.LastPrice),
	}).Debug("Received Event")
	if isNeedAlert(newEvent) {
		if err := PostMessageToTgChannel(helper.GetRequiredStringEnv("TELEGRAM_CHANNEL_USERNAME"), newEvent.AlertText()); err != nil {
			log.WithField("Error", err).Error("Post message to tg channel failed")
		}
	}
	lastAlert[newEvent.Symbol] = newEvent
}

func errHandler(err error) {
	log.Error(err)
}

func init() {
	binance.WebsocketKeepalive = true
}

func main() {
	symbols := lo.Map[string, string](baseAssets, func(baseAsset string, i int) string {
		return baseAsset + quoteAsset
	})
	for {
		log.Debug("Connect to binance...")
		doneC, _, err := binance.WsCombinedMarketStatServe(symbols, eventHandler, errHandler)
		if err != nil {
			log.Error(err)
			log.Debug("Connect Failed, Reconnect in 3 seconds")
			time.Sleep(time.Second * 3)
			continue
		}
		<-doneC
		log.Debug("Disconnected, Reconnect in 3 seconds")
		time.Sleep(time.Second * 3)
	}
}
