package minds

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type ThreadContext interface {
	Context() context.Context
	UUID() string
	Messages() Messages
	Metadata() Metadata

	AppendMessages(message ...Message) ThreadContext
	WithContext(ctx context.Context) ThreadContext
	WithUUID(uuid string) ThreadContext
	WithMessages(message ...Message) ThreadContext
	WithMetadata(metadata Metadata) ThreadContext
}

type threadContext struct {
	mu       sync.RWMutex
	ctx      context.Context
	uuid     string
	metadata Metadata
	messages Messages
}

func NewThreadContext(ctx context.Context) ThreadContext {
	return &threadContext{
		ctx:      ctx,
		uuid:     uuid.New().String(),
		metadata: Metadata{},
		messages: Messages{},
	}
}

func (tc *threadContext) AppendMessages(messages ...Message) ThreadContext {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.messages = append(tc.messages, messages...)
	return tc
}

func (tc *threadContext) Context() context.Context {
	return tc.ctx
}

func (tc *threadContext) UUID() string {
	return tc.uuid
}

func (tc *threadContext) Messages() Messages {
	return tc.messages.Copy()
}

func (tc *threadContext) Metadata() Metadata {
	return tc.metadata.Copy()
}

func (tc *threadContext) WithContext(ctx context.Context) ThreadContext {
	return &threadContext{
		ctx:      ctx,
		uuid:     tc.UUID(),
		metadata: tc.metadata.Copy(),
		messages: tc.messages.Copy(),
	}
}

func (tc *threadContext) WithUUID(uuid string) ThreadContext {
	return &threadContext{
		ctx:      tc.Context(),
		uuid:     uuid,
		metadata: tc.metadata.Copy(),
		messages: tc.messages.Copy(),
	}
}

func (tc *threadContext) WithMessages(messages ...Message) ThreadContext {
	return &threadContext{
		ctx:      tc.Context(),
		uuid:     tc.UUID(),
		metadata: tc.metadata.Copy(),
		messages: messages,
	}
}

func (tc *threadContext) WithMetadata(metadata Metadata) ThreadContext {
	return &threadContext{
		ctx:      tc.Context(),
		uuid:     tc.UUID(),
		metadata: metadata,
		messages: tc.messages.Copy(),
	}
}
