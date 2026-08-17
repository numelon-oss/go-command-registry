package commandregistry

import (
	"context"
	"flag"
)

type Handler[Environment any] func(ctx context.Context, env Environment, args []string) error

type Configure[Environment any] func(flags *flag.FlagSet) Handler[Environment]

type Command[Environment any] struct {
	Name        string
	Description string
	Arguments   string
	Example     string

	Available func(env Environment) bool
	Configure Configure[Environment]

	children map[string]*Command[Environment]
}

func (command *Command[Environment]) Add(children ...*Command[Environment]) {
	if command.children == nil {
		command.children = make(map[string]*Command[Environment])
	}
	for _, child := range children {
		if child == nil {
			panic("nil command")
		}
		if child.Name == "" {
			panic("empty command name")
		}
		if _, exists := command.children[child.Name]; exists {
			panic("duplicate command")
		}
		command.children[child.Name] = child
	}
}

func (command *Command[Environment]) child(name string) (*Command[Environment], bool) {
	child, exists := command.children[name]
	return child, exists
}

func (command *Command[Environment]) available(env Environment) bool {
	return command.Available == nil || command.Available(env)
}

func Simple[Environment any](handler Handler[Environment]) Configure[Environment] {
	return func(_ *flag.FlagSet) Handler[Environment] {
		return handler
	}
}
