package commandregistry

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

func (registry *Registry[Environment]) PrintHelp(
	env Environment,
	command *Command[Environment],
	path []string,
) {
	output := io.Discard
	if registry.config.Output != nil {
		if configured := registry.config.Output(env); configured != nil {
			output = configured
		}
	}

	writer := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	defer writer.Flush()

	fmt.Fprintln(writer, "Usage:")
	fmt.Fprintf(writer, "  %s", registry.config.Program)
	if command == registry.root {
		fmt.Fprintf(writer, " %s", registry.config.RootArguments)
	} else {
		fmt.Fprintf(writer, " %s", strings.Join(path, " "))
		if command.Arguments != "" {
			fmt.Fprintf(writer, " %s", command.Arguments)
		}
	}
	fmt.Fprintln(writer)

	if command.Description != "" {
		fmt.Fprintf(writer, "\n%s\n", command.Description)
	}

	children := visibleChildren(command, env)
	if command == registry.root || len(children) > 0 {
		fmt.Fprintln(writer, "\nCommands:")
		if command == registry.root {
			fmt.Fprintln(writer, "  help\tShow command help")
		}
		for _, child := range children {
			fmt.Fprintf(writer, "  %s\t%s\n", child.Name, child.Description)
		}
	}

	if command == registry.root && len(registry.config.GlobalOptions) > 0 {
		fmt.Fprintln(writer, "\nGlobal options:")
		for _, option := range registry.config.GlobalOptions {
			fmt.Fprintf(writer, "  %s\t%s\n", option.Usage, option.Description)
		}
	} else if command != registry.root && command.Configure != nil {
		flags := flag.NewFlagSet(command.Name, flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		command.Configure(flags)
		if hasFlags(flags) {
			fmt.Fprintln(writer, "\nOptions:")
			flags.VisitAll(func(option *flag.Flag) {
				fmt.Fprintf(writer, "  --%s\t%s\n", option.Name, option.Usage)
			})
		}
	}

	if command.Example != "" {
		fmt.Fprintln(writer, "\nExample:")
		fmt.Fprintf(writer, "  %s\n", command.Example)
	}
}

func hasFlags(flags *flag.FlagSet) bool {
	hasAny := false
	flags.VisitAll(func(_ *flag.Flag) {
		hasAny = true
	})
	return hasAny
}

func visibleChildren[Environment any](
	command *Command[Environment],
	env Environment,
) []*Command[Environment] {
	children := make([]*Command[Environment], 0, len(command.children))
	for _, child := range command.children {
		if child.available(env) {
			children = append(children, child)
		}
	}
	sort.Slice(children, func(left int, right int) bool {
		return children[left].Name < children[right].Name
	})
	return children
}
