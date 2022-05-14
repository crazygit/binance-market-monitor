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
	LastPriceFloat          float64
	CloseQtyFloat           float64
}

func escapeTextToMarkdownV2(text string) string {
	return tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, text)
}

func getDifference(newValue, oldValue float64, suffix string) string {
	diff := newValue - oldValue
	direction := "🔻"
	if diff > 0 {
		direction = "🔺"
	}
	return fmt.Sprintf("(%s %f%s)", direction, diff, suffix)
}

func (newEvent ExtendWsMarketStatEvent) AlertText(oldEvent ExtendWsMarketStatEvent) string {
	return fmt.Sprintf(`
*交易对*: %s

_最新报警信息_
*最新成交价格*: %s
*24小时价格变化百分比*: %s
*最新成交价格上的成交量*: %s
*24小时内成交量*: %s
*24小时内成交额*: %s

_上次报警信息_

*上次报警价格*: %s
*上次价格变化百分比*: %s
*上次价格上的成交量*: %s

两次报警间隔时间: %s
`, escapeTextToMarkdownV2(prettySymbol(newEvent.Symbol)),

		escapeTextToMarkdownV2("$"+prettyFloatString(newEvent.LastPrice)),          // 最新成交价格
		escapeTextToMarkdownV2(prettyFloatString(newEvent.PriceChangePercent)+"%"), //  24小时价格变化(百分比)
		escapeTextToMarkdownV2(prettyFloatString(newEvent.CloseQty)),               // 最新成交价格上的成交量
		escapeTextToMarkdownV2(prettyFloatString(newEvent.BaseVolume)),             // 24小时内成交量
		escapeTextToMarkdownV2(prettyFloatString(newEvent.QuoteVolume)),            // 24小时内成交额

		escapeTextToMarkdownV2("$"+prettyFloatString(oldEvent.LastPrice)+" "+getDifference(newEvent.LastPriceFloat, oldEvent.LastPriceFloat, "")),                             //上次报警价格
		escapeTextToMarkdownV2(prettyFloatString(oldEvent.PriceChangePercent)+"%"+" "+getDifference(newEvent.PriceChangePercentFloat, oldEvent.PriceChangePercentFloat, "%")), //上次价格变化百分比
		escapeTextToMarkdownV2("$"+prettyFloatString(oldEvent.CloseQty)+" "+getDifference(newEvent.CloseQtyFloat, oldEvent.CloseQtyFloat, "")),                                //上次价格上的成交量

		escapeTextToMarkdownV2(time.UnixMilli(newEvent.Time).Truncate(time.Second).Sub(time.UnixMilli(oldEvent.Time).Truncate(time.Second)).String()), //两次报警间隔时间
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
		// 首次启动时会触发大量报警，忽略程序启动时,波动已经大于预设值的报警
		//return math.Abs(newEvent.PriceChangePercentFloat) >= priceChangePercentThreshold
		lastAlert[newEvent.Symbol] = newEvent
		return false
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
	log.WithFields(logrus.Fields{"SymbolsInAlertMap": len(lastAlert), "RevivedEventsNumber": len(events)}).Info("Stats")
	for _, event := range events {
		if isIgnoreEvent(event) {
			continue
		}
		priceChangePercentFloat, _ := strconv.ParseFloat(event.PriceChangePercent, 64)
		lastPriceFloat, _ := strconv.ParseFloat(event.LastPrice, 64)
		closeQtyFloat, _ := strconv.ParseFloat(event.CloseQty, 64)
		newEvent := ExtendWsMarketStatEvent{
			WsMarketStatEvent:       event,
			PriceChangePercentFloat: priceChangePercentFloat,
			LastPriceFloat:          lastPriceFloat,
			CloseQtyFloat:           closeQtyFloat,
		}
		log.WithFields(logrus.Fields{
			"Symbol":             newEvent.Symbol,
			"PriceChange":        prettyFloatString(newEvent.LastPrice),
			"PriceChangePercent": newEvent.PriceChangePercent,
			"LastPrice":          prettyFloatString(newEvent.LastPrice),
			"Time":               newEvent.Time,
		}).Debug("Received Event")
		if isNeedAlert(newEvent) {
			postMessageTextBuilder.WriteString(newEvent.AlertText(lastAlert[newEvent.Symbol]))
			lastAlert[newEvent.Symbol] = newEvent
			postMessage = true
		}
	}
	if postMessage {
		postMessageTextBuilder.WriteString(fmt.Sprintf("\n\n%s", escapeTextToMarkdownV2(fmt.Sprintf("(%s)", time.Now().Format(time.RFC3339)))))
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
