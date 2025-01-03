package minds

import (
	"errors"
)

var (
	ErrNoMessages = errors.New("no messages in thread")
)

type Messages []Message

// Copy returns a deep copy of the messages
func (m Messages) Copy() Messages {
	copied := make(Messages, len(m))
	for i, msg := range m {
		newMsg := Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			Metadata:   msg.Metadata.Copy(),
			ToolCallID: msg.ToolCallID,
			ToolCalls:  msg.ToolCalls,
		}
		copied[i] = newMsg
	}
	return copied
}

// Last returns the last message in the slice of messages. NOTE: This will
// return an empty message if there are no messages in the slice.
func (m Messages) Last() Message {
	if len(m) == 0 {
		return Message{}
	}

	return m[len(m)-1]
}

// Exclude returns a new slice of Message with messages with the specified roles removed
func (m Messages) Exclude(roles ...Role) Messages {
	filteredMsgs := Messages{}

	for _, msg := range m {
		keep := false
		for _, role := range roles {
			if msg.Role == role {
				continue
			}
			keep = true
		}

		if keep {
			filteredMsgs = append(filteredMsgs, msg)
		}
	}

	return filteredMsgs
}

func (m Messages) Only(roles ...Role) Messages {
	filteredMsgs := Messages{}

	for _, msg := range m {
		for _, role := range roles {
			if msg.Role == role {
				filteredMsgs = append(filteredMsgs, msg)
				break
			}
		}
	}

	return filteredMsgs
}

func (m Messages) TokenCount(tokenizer TokenCounter) (int, error) {
	total := 0
	for _, msg := range m {
		count, err := tokenizer.CountTokens(msg.Content)
		if err != nil {
			return 0, err
		}

		total += count
	}

	return total, nil
}
