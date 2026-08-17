package commandregistry

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"
)

type Result int

const (
	ResultSuccess Result = iota
	ResultFailure
	ResultUsage
)

type Option struct {
	Usage       string
	Description string
}

type Config[Environment any] struct {
	Program         string
	RootArguments   string
	RootDescription string
	GlobalOptions   []Option
	EmptyResult     Result

	Output        func(env Environment) io.Writer
	ReportError   func(env Environment, err error)
	ClassifyError func(err error) Result
}

type Registry[Environment any] struct {
	config Config[Environment]
	root   *Command[Environment]
}

func New[Environment any](config Config[Environment]) *Registry[Environment] {
	if config.Program == "" {
		panic("empty program name")
	}
	if config.RootArguments == "" {
		config.RootArguments = "<command> [arguments]"
	}
	return &Registry[Environment]{
		config: config,
		root: &Command[Environment]{
			Name:        config.Program,
			Description: config.RootDescription,
		},
	}
}

func (registry *Registry[Environment]) Register(commands ...*Command[Environment]) {
	registry.root.Add(commands...)
}

func (registry *Registry[Environment]) Execute(
	ctx context.Context,
	env Environment,
	args []string,
) Result {
	if len(args) == 0 {
		registry.PrintHelp(env, registry.root, nil)
		return registry.config.EmptyResult
	}
	if args[0] == "help" {
		return registry.executeHelp(env, args[1:])
	}

	command, path, remaining, err := registry.resolve(env, args)
	if err != nil {
		return registry.finish(env, command, path, err)
	}
	if containsHelp(remaining) {
		registry.PrintHelp(env, command, path)
		return ResultSuccess
	}
	if command.Configure == nil {
		if len(remaining) > 0 {
			return registry.finish(
				env,
				command,
				path,
				UsageErrorf("Unexpected argument %q", remaining[0]),
			)
		}
		registry.PrintHelp(env, command, path)
		return ResultSuccess
	}

	flags := flag.NewFlagSet(strings.Join(path, " "), flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	handler := command.Configure(flags)
	if err := flags.Parse(remaining); err != nil {
		return registry.finish(env, command, path, &UsageError{Message: err.Error()})
	}

	return registry.finish(env, command, path, handler(ctx, env, flags.Args()))
}

func (registry *Registry[Environment]) resolve(
	env Environment,
	args []string,
) (*Command[Environment], []string, []string, error) {
	command := registry.root
	path := make([]string, 0, 3)

	for index, argument := range args {
		if strings.HasPrefix(argument, "-") {
			return command, path, args[index:], nil
		}
		child, exists := command.child(argument)
		if !exists || !child.available(env) {
			if command == registry.root || command.Configure == nil {
				return command, path, nil, UsageErrorf("Unknown command %q", argument)
			}
			return command, path, args[index:], nil
		}

		command = child
		path = append(path, argument)
		if index == len(args)-1 {
			return command, path, nil, nil
		}
	}

	return command, path, nil, nil
}

func (registry *Registry[Environment]) executeHelp(env Environment, path []string) Result {
	command := registry.root
	resolved := make([]string, 0, len(path))
	for _, name := range path {
		child, exists := command.child(name)
		if !exists || !child.available(env) {
			return registry.finish(env, command, resolved, UsageErrorf("Unknown command %q", name))
		}
		command = child
		resolved = append(resolved, name)
	}

	registry.PrintHelp(env, command, resolved)
	return ResultSuccess
}

func (registry *Registry[Environment]) finish(
	env Environment,
	command *Command[Environment],
	path []string,
	err error,
) Result {
	if err == nil {
		return ResultSuccess
	}

	result := ResultFailure
	var usageError *UsageError
	if errors.As(err, &usageError) {
		result = ResultUsage
	} else if registry.config.ClassifyError != nil {
		result = registry.config.ClassifyError(err)
	}

	if registry.config.ReportError != nil {
		registry.config.ReportError(env, err)
	}
	if result == ResultUsage {
		registry.PrintHelp(env, command, path)
	}
	return result
}

func containsHelp(args []string) bool {
	for _, argument := range args {
		if argument == "--help" || argument == "-h" {
			return true
		}
	}
	return false
}
