package main

import (
	"fmt"

	"github.com/BotMatrix/src/plugins/sdk"
)

func main() {
	p := sdk.NewPlugin()

	p.OnMessage(func(msg *sdk.EventMessage) ([]sdk.Action, error) {
		payload, err := msg.GetPayload()
		if err != nil {
			return nil, err
		}

		text, textOk := payload["text"].(string)
		target, targetOk := payload["from"].(string)
		targetID, targetIDOk := payload["group_id"].(string)

		if !textOk || !targetOk || !targetIDOk {
			return nil, fmt.Errorf("missing required fields in payload")
		}

		// TODO: Add plugin logic here
		return []sdk.Action{
			{
				Type:     "send_message",
				Target:   target,
				TargetID: targetID,
				Text:     fmt.Sprintf("This is a placeholder response from badge plugin, received: %s", text),
			},
		}, nil
	})

	p.Run()
}
