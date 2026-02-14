# othelp

Thin OpenTelemetry instrumentation helper for Go — less boilerplate, same power.

## Problem

OTel instrumentation in Go is verbose. Every function needs manual error recording, status setting, and span cleanup:

```go
func GetUser(ctx context.Context, id string) (*User, error) {
    ctx, span := otel.Tracer("myapp").Start(ctx, "GetUser")
    defer span.End()
    span.SetAttributes(attribute.String("user.id", id))

    user, err := db.FindUser(ctx, id)
    if err != nil {
        span.RecordError(err)                    // easy to forget
        span.SetStatus(codes.Error, err.Error())  // easy to forget
        return nil, err
    }
    span.SetStatus(codes.Ok, "")
    return user, nil
}
```

Forgetting `RecordError` or `SetStatus` is the most common OTel instrumentation bug.

## Solution

`othelp` provides a `defer end(&err)` pattern that handles error recording automatically:

```go
func GetUser(ctx context.Context, id string) (user *User, err error) {
    ctx, end := tracer.Start(ctx, "GetUser",
        othelp.Str("user.id", id),
    )
    defer end(&err) // automatically records error and sets status

    user, err = db.FindUser(ctx, id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

## Install

```bash
go get github.com/a1yama/othelp
```

## Usage

### Initialization

```go
shutdown, err := othelp.Init(ctx, othelp.Config{
    ServiceName: "myapp",
    Exporter:    "otlp",       // "otlp" or "stdout"
    Endpoint:    "localhost:4317",
    Insecure:    true,
})
if err != nil {
    log.Fatal(err)
}
defer shutdown(ctx)
```

### Creating a Tracer

```go
var tracer = othelp.NewTracer("myapp/usecase")
```

### Instrumenting Functions

```go
func CreateOrder(ctx context.Context, req OrderRequest) (order *Order, err error) {
    ctx, end := tracer.Start(ctx, "CreateOrder",
        othelp.Str("order.type", req.Type),
        othelp.Int("order.items", len(req.Items)),
    )
    defer end(&err)

    // your logic here — just return errors normally
    order, err = db.InsertOrder(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("insert order: %w", err)
    }
    return order, nil
}
```

### Attribute Helpers

```go
othelp.Str("key", "value")
othelp.Int("key", 42)
othelp.Int64("key", int64(42))
othelp.Float64("key", 3.14)
othelp.Bool("key", true)
othelp.Strs("key", []string{"a", "b"})
othelp.Ints("key", []int{1, 2, 3})
```

### Escape Hatch

Access the underlying OTel tracer when you need full control:

```go
otelTracer := tracer.OTelTracer()
```

## Design Principles

1. **`defer end(&err)` is the core** — structurally eliminates the most common instrumentation bug
2. **Thin wrapper** — never hides the OTel API; always provides an escape hatch
3. **Minimal dependencies** — only depends on the official OTel SDK

## License

MIT
