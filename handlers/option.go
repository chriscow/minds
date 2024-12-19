package handlers

import "github.com/chriscow/minds"

type HandlerOption struct {
	name        string
	description string
	prompt      minds.Prompt
	// handler     minds.ThreadHandler
}

func WithName(name string) func(*HandlerOption) {
	return func(ho *HandlerOption) {
		ho.name = name
	}
}

func WithDescription(description string) func(*HandlerOption) {
	return func(ho *HandlerOption) {
		ho.description = description
	}
}

func WithPrompt(prompt minds.Prompt) func(*HandlerOption) {
	return func(ho *HandlerOption) {
		ho.prompt = prompt
	}
}
