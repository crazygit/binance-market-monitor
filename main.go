package main

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/crazygit/binance-market-monitor/helper"
	l "github.com/crazygit/binance-market-monitor/helper/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"math"
	"strconv"
	"strings"
	"time"
)

var log = l.GetLog()

var (
	quoteAsset                  = strings.ToUpper(helper.GetStringEnv("QUOTE_ASSET", "USDT"))
	lowestPriceFilter           = helper.GetFloat64Env("LOWEST_PRICE_FILTER", 1.0)
	priceChangePercentThreshold = helper.GetFloat64Env("PRICE_CHANGE_PERCENT_THRESHOLD", 5.0)
)

var lastAlert = map[string]ExtendWsMarketStatEvent{}

type ExtendWsMarketStatEvent struct {
	*binance.WsMarketStatEvent
	PriceChangePercentFloat float64
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
	replacer = strings.NewReplacer(quoteAsset, fmt.Sprintf("/%s", quoteAsset))
	return replacer.Replace(symbol)
}

func isNeedAlert(newEvent ExtendWsMarketStatEvent) bool {
	if oldEvent, ok := lastAlert[newEvent.Symbol]; ok {
		return math.Abs(oldEvent.PriceChangePercentFloat-newEvent.PriceChangePercentFloat) >= priceChangePercentThreshold
	} else {
		return math.Abs(newEvent.PriceChangePercentFloat) >= priceChangePercentThreshold
	}
}

func isIgnoreEvent(event *binance.WsMarketStatEvent) bool {
	weightedAvgPrice, _ := strconv.ParseFloat(event.WeightedAvgPrice, 64)
	if !strings.HasSuffix(event.Symbol, quoteAsset) || weightedAvgPrice < lowestPriceFilter {
		return true
	}
	return false
}

func eventHandler(events binance.WsAllMarketsStatEvent) {
	var postMessageTextBuilder strings.Builder
	var postMessage = false
	for _, event := range events {
		if isIgnoreEvent(event) {
			continue
		}
		priceChangePercentFloat, _ := strconv.ParseFloat(event.PriceChangePercent, 64)
		newEvent := ExtendWsMarketStatEvent{WsMarketStatEvent: event, PriceChangePercentFloat: priceChangePercentFloat}
		log.WithFields(logrus.Fields{
			"Symbol":             newEvent.Symbol,
			"PriceChange":        prettyFloatString(newEvent.LastPrice),
			"PriceChangePercent": newEvent.PriceChangePercent,
			"LastPrice":          prettyFloatString(newEvent.LastPrice),
			"Time":               newEvent.Time,
		}).Debug("Received Event")
		if isNeedAlert(newEvent) {
			postMessageTextBuilder.WriteString(newEvent.AlertText())
			lastAlert[newEvent.Symbol] = newEvent
			postMessage = true
		}
	}
	if postMessage {
		postMessageTextBuilder.WriteString(fmt.Sprintf("\n\n%s", escapeTextToMarkdownV2(fmt.Sprintf("(%s)", time.Now()))))
		if err := PostMessageToTgChannel(getTelegramChannelName(), postMessageTextBuilder.String()); err != nil {
			log.WithField("Error", err).Error("Post message to tg channel failed")
		}
	}
}

func getTelegramChannelName() string {
	channelName := helper.GetRequiredStringEnv("TELEGRAM_CHANNEL_USERNAME")
	if !strings.HasPrefix(channelName, "@") {
		return "@" + channelName
	}
	return channelName
}

func errHandler(err error) {
	log.Error(err)
}

func init() {
	binance.WebsocketKeepalive = true
}

func main() {
	for {
		log.Info("Connect to binance...")
		doneC, _, err := binance.WsAllMarketsStatServe(eventHandler, errHandler)
		if err != nil {
			log.Error(err)
			continue
		}
		<-doneC
	}
}
