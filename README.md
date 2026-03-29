# ldi
Lightweight dependency injection

__not thread safe__

## Features

- Dependency injection with function and value providers
- Parent-child container support
- Circular dependency detection
- Memory leak prevention with automatic cleanup
- Resource management with Clear() and CleanupResolutionTracking() methods

## Usage

```go
di := New()
di.Provide("config value")
di.Provide(func(cfg string) Database {
    return NewDatabase(cfg)
})

di.Invoke(func(db Database) {
    // use database
})
```

## Memory Management

The library automatically cleans up resolution tracking to prevent memory leaks. 
For manual cleanup, use:

- `Clear()` - Removes all providers and resets container state
- `CleanupResolutionTracking()` - Clears resolution tracking entries
