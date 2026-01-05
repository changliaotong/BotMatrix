module consolebot

go 1.25.0

require (
	BotMatrix/common v0.0.0
	github.com/gorilla/websocket v1.5.3
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
)

replace BotMatrix/common => ../Common
