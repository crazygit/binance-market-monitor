# 币安市场行情监控

## 币安文档要点

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

### 连接示例

```shell
> websocat wss://stream.binance.com:9443/ws/btcusdt@ticker
```
