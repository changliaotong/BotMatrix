package mobile

import (
	"botmatrix/wxbotgo/core"
)

// Callback interface exported to Java/Dart
type Callback interface {
	OnLog(msg string)
	OnQrCode(url string)
}

// Internal wrapper to adapt mobile.Callback to core.BotCallback
type mobileCallbackWrapper struct {
	cb Callback
}

func (w *mobileCallbackWrapper) OnLog(msg string) {
	if w.cb != nil {
		w.cb.OnLog(msg)
	}
}

func (w *mobileCallbackWrapper) OnQrCode(url string) {
	if w.cb != nil {
		w.cb.OnQrCode(url)
	}
}

// Start the bot.
func Start(managerUrl, selfId string, cb Callback) {
	bot := core.NewWxBot(managerUrl, selfId, &mobileCallbackWrapper{cb: cb})
	bot.Start()
}
