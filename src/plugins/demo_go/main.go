package main

import (
	"BotMatrix/src/plugins/sdk"
	"fmt"
	"strings"
)

func main() {
	p := sdk.NewPlugin()

	p.OnMessage(func(ctx *sdk.Context) error {
		text := ctx.Event.Payload["text"].(string)
		if strings.HasPrefix(text, "/echo ") {
			content := strings.TrimPrefix(text, "/echo ")
			ctx.Reply(fmt.Sprintf("Go SDK Echo: %s", content))
		}
		return nil
	})

	p.Run()
}
