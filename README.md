# 币安市场行情监控

监控币安市场24小时内现货行情波动情况，当24小时内币价波动的百分比超过设置的百分数后，自动发送警报到Telegram Channel

## 触发报警的条件

经过大致分析了币安现货市场价格波动后，为了抓住有效的价格波动，排除干扰信息。根据币价的不同，设置了不同的价格波动百分比报警

- 币价`>= 300$`, 价格波动 `>=6%` 发生报警
- `$1` < 币价 < `$300` 时，波动超过 `>=15%` 发生报警
- 币价 `<= 1$`, 波动 `>=50%` 发生报警

对于同一个币种`10分钟`内只报警一次

### 配置环境变量

## 现成版

订阅 [BinanceMarketMonitor](https://t.me/BinanceMarketMonitor) 即可

## 本地运行或定制

### 环境变量作用

- `TELEGRAM_API_TOKEN` Telegram 机器的人的API Token
  开通方式可参考官方文档[How do I create a bot?](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- `TELEGRAM_CHANNEL_USERNAME` Telegram Channel的名字
- `QUOTE_ASSET` 交易对的定价资产，默认为`USDT`。只会监控所有以`USDT`作为定价资产的交易对价格波动。可以修改为`BTC`, `BNB`,`BUSDT`等其它币安支持的定价资产
- `ENVIRONMENT` 运行环境，可选值为`dev`或`production`，区别在于当设置为`dev`时，运行时会输出更多的log信息

设置本地环境变量

```shell
$ git clone https://github.com/crazygit/binance-market-monitor.git
$ cd binance-market-monitor
$ cp .env.example .env
```

然后根据[环境变量作用](#环境变量作用)的介绍修改`.env`文件

### 启动服务

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
