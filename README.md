# 币安市场行情监控

监控币安市场现货行情24内行情波动情况，当24小时类价格波动的百分比超过设置的百分数后，自动发送警报到Telegram Channel

## 现成版

订阅 Telegram [BinanceMarketMonitor](https://t.me/BinanceMarketMonitor) Channel即可

## 环境变量作用

- `TELEGRAM_API_TOKEN` Telegram 机器的人API Token,
  开通方式可参考官方文档[How do I create a bot?](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- `TELEGRAM_CHANNEL_USERNAME` Telegram Channel的名字
- `QUOTE_ASSET` 交易对的定价资产，默认为`USDT`。只会监控所有以`USDT`作为定价资产交易对的价格变化波动。可以修改为`BTC`, `BNB`,`BUSDT`等其它币安支持的定价资产
- `LOWEST_PRICE_FILTER` 交易对象的最低价格过滤，默认为`1.0`。价格低的币种行情波动大，没有参考价值，通过该配置，可以自动过滤掉24小时内平均价格低于该配置的交易对信息
- `PRICE_CHANGE_PERCENT_THRESHOLD` 触发报警的价格波动百分比，默认为`5.0`。当24小时内价格波动百分大于该值时，触发报警
- `ENVIRONMENT` 运行环境，可选值为`dev`或`production`，区别在于当设置为`dev`时，运行时会输出更多的log信息

## 本地运行

设置本地环境变量

```shell
$ copy .env.example .env
```

然后根据[环境变量作用](#环境变量作用)的介绍修改`.env`文件

启动服务

```shell
$ docker compose up
```

## 其它

### 币安文档要点

[接口文档](https://binance-docs.github.io/apidocs/spot/cn/)

### Websocket 行情推送

- 本篇所列出的所有`wss`接口的`baseurl`为: `wss://stream.binance.com:9443`
- `Streams`有单一原始`stream`或组合`stream`
- 单一原始`streams` 格式为 `/ws/<streamName>`
- 组合streams的URL格式为 `/stream?streams=<streamName1>/<streamName2>/<streamName3>`
- 订阅组合streams时，事件payload会以这样的格式封装: `{"stream":"<streamName>","data":<rawPayload>}`
- `stream`名称中所有交易对均为 小写
- 每个到`stream.binance.com`的链接有效期不超过24小时，请妥善处理断线重连。
- 每3分钟，服务端会发送`ping`帧，客户端应当在10分钟内回复`pong`帧，否则服务端会主动断开链接。允许客户端发送不成对的`pong`帧(即客户端可以以高于10分钟每次的频率发送`pong`帧保持链接)。

### 术语

- `base asset`指一个交易对的交易对象，即写在靠前部分的资产名, 比如`BTCUSDT`, `BTC`是`base asset`。
- `quote asset`指一个交易对的定价资产，即写在靠后部分的资产名, 比如`BTCUSDT`, `USDT`是`quote asset`。

### 命令行连接示例

```shell
> websocat wss://stream.binance.com:9443/ws/btcusdt@ticker
```
