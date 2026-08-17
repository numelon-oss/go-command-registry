# go-command-registry

This is a small hierarchical command registry for command-line applications.

It provides nested help, fresh command flags for every execution, application-defined command availability, usage errors and configurable result and error handling.

## Integration

An application can bind its environment type once with local aliases, keeping command declarations concise without giving up type safety:

```go
type Handler = commandregistry.Handler[*Environment]
type Configure = commandregistry.Configure[*Environment]
type Command = commandregistry.Command[*Environment]
type Registry = commandregistry.Registry[*Environment]
```

## Example

The [`example`](./example) directory contains a complete runnable command-line application showing nested commands, flags, application-owned state, usage errors and conditional command availability.

```sh
go run ./example help
go run ./example help service create
go run ./example service list
go run ./example service create --name worker --kind service
go run ./example --admin admin status
```
