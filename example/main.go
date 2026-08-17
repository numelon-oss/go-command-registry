package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	commandregistry "github.com/numelon-oss/go-command-registry"
)

type Service struct {
	ID   int
	Name string
	Kind string
}

type Environment struct {
	Output      io.Writer
	ErrorOutput io.Writer
	Services    []Service
	Admin       bool
}

type Handler = commandregistry.Handler[*Environment]
type Configure = commandregistry.Configure[*Environment]
type Command = commandregistry.Command[*Environment]
type Registry = commandregistry.Registry[*Environment]

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	global := flag.NewFlagSet("deployctl", flag.ContinueOnError)
	global.SetOutput(io.Discard)
	admin := global.Bool("admin", false, "Enable administrator commands")
	help := global.Bool("help", false, "Show command help")
	shortHelp := global.Bool("h", false, "Show command help")
	if err := global.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return int(commandregistry.ResultUsage)
	}

	env := &Environment{
		Output:      os.Stdout,
		ErrorOutput: os.Stderr,
		Admin:       *admin,
		Services: []Service{
			{ID: 1, Name: "api", Kind: "service"},
			{ID: 2, Name: "website", Kind: "static"},
		},
	}

	args = global.Args()
	if *help || *shortHelp {
		args = append([]string{"help"}, args...)
	}

	registry := newRegistry()
	return int(registry.Execute(context.Background(), env, args))
}

func newRegistry() *Registry {
	registry := commandregistry.New(commandregistry.Config[*Environment]{
		Program:         "deployctl",
		RootArguments:   "[global options] <command> [arguments]",
		RootDescription: "A small deployment CLI thing test",
		EmptyResult:     commandregistry.ResultUsage,
		GlobalOptions: []commandregistry.Option{
			{Usage: "--admin", Description: "Enable administrator commands"},
			{Usage: "--help, -h", Description: "Show command help"},
		},
		Output: func(env *Environment) io.Writer {
			return env.Output
		},
		ReportError: func(env *Environment, err error) {
			fmt.Fprintf(env.ErrorOutput, "error: %v\n", err)
		},
	})

	registry.Register(serviceCommand(), adminCommand())
	return registry
}

func serviceCommand() *Command {
	command := &Command{
		Name:        "service",
		Description: "Manage services",
	}
	command.Add(listServicesCommand(), showServiceCommand(), createServiceCommand())
	return command
}

func listServicesCommand() *Command {
	return &Command{
		Name:        "list",
		Description: "List services",
		Configure: commandregistry.Simple(func(
			_ context.Context,
			env *Environment,
			args []string,
		) error {
			if len(args) != 0 {
				return commandregistry.UsageErrorf("Service list does not accept arguments")
			}

			fmt.Fprintln(env.Output, "ID\tNAME\tKIND")
			for _, service := range env.Services {
				fmt.Fprintf(env.Output, "%d\t%s\t%s\n", service.ID, service.Name, service.Kind)
			}
			return nil
		}),
	}
}

func showServiceCommand() *Command {
	return &Command{
		Name:        "show",
		Description: "Show a service",
		Arguments:   "<id>",
		Example:     "deployctl service show 1",
		Configure: commandregistry.Simple(func(
			_ context.Context,
			env *Environment,
			args []string,
		) error {
			if len(args) != 1 {
				return commandregistry.UsageErrorf("Expected one service ID")
			}

			id, err := strconv.Atoi(args[0])
			if err != nil || id < 1 {
				return commandregistry.UsageErrorf("Service ID must be a positive integer")
			}
			for _, service := range env.Services {
				if service.ID == id {
					fmt.Fprintf(env.Output, "ID: %d\nName: %s\nKind: %s\n", service.ID, service.Name, service.Kind)
					return nil
				}
			}
			return fmt.Errorf("service %d does not exist", id)
		}),
	}
}

func createServiceCommand() *Command {
	return &Command{
		Name:        "create",
		Description: "Create a service",
		Arguments:   "--name <name> --kind <kind>",
		Example:     "deployctl service create --name worker --kind service",
		Configure: func(flags *flag.FlagSet) Handler {
			name := flags.String("name", "", "Service name")
			kind := flags.String("kind", "", "Service kind: service or static")

			return func(_ context.Context, env *Environment, args []string) error {
				if len(args) != 0 {
					return commandregistry.UsageErrorf("Service create does not accept positional arguments")
				}
				if strings.TrimSpace(*name) == "" {
					return commandregistry.UsageErrorf("Service name is required")
				}
				if *kind != "service" && *kind != "static" {
					return commandregistry.UsageErrorf("Service kind must be service or static")
				}

				service := Service{ID: len(env.Services) + 1, Name: *name, Kind: *kind}
				env.Services = append(env.Services, service)
				fmt.Fprintf(env.Output, "Created service %d (%s)\n", service.ID, service.Name)
				return nil
			}
		},
	}
}

func adminCommand() *Command {
	command := &Command{
		Name:        "admin",
		Description: "Inspect administrative state",
		Available: func(env *Environment) bool {
			return env.Admin
		},
	}
	command.Add(&Command{
		Name:        "status",
		Description: "Show administrative status",
		Configure: commandregistry.Simple(func(
			_ context.Context,
			env *Environment,
			args []string,
		) error {
			if len(args) != 0 {
				return commandregistry.UsageErrorf("Admin status does not accept arguments")
			}
			fmt.Fprintf(env.Output, "Administrator access enabled for %d services\n", len(env.Services))
			return nil
		}),
	})
	return command
}
