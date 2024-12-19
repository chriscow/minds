package minds

type ThreadHandler interface {
	HandleThread(thread ThreadContext, next ThreadHandler) (ThreadContext, error)
}

type ThreadHandlerFunc func(thread ThreadContext, next ThreadHandler) (ThreadContext, error)

func (f ThreadHandlerFunc) HandleThread(thread ThreadContext, next ThreadHandler) (ThreadContext, error) {
	return f(thread, next)
}

type NoopThreadHandler struct{}

func (h NoopThreadHandler) HandleThread(tc ThreadContext, next ThreadHandler) (ThreadContext, error) {
	return tc, nil
}
