package tools

import (
	"context"
	"encoding/json"

	"github.com/chriscow/minds"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// NewToolNode creates a handler function for processing LLM requests
func NewToolNode(ctx context.Context, tool minds.Tool, logger watermill.LoggerAdapter) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {

		result, err := tool.Call(ctx, msg.Payload)
		if err != nil {
			return nil, err
		}

		responseMsg := message.NewMessage(
			watermill.NewUUID(),
			[]byte(result),
		)

		if correlationID, ok := msg.Metadata["correlation_id"]; ok {
			responseMsg.Metadata["correlation_id"] = correlationID
		}

		return []*message.Message{responseMsg}, nil
	}
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
