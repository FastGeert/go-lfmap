# Lock free threadsafe hashmap written in pure Go

I needed a map implementation that is safe for concurrent access. The current implementation is done via channels, which probably can be improved by using a normal simple mutex.

## Installation 

```bash
go get github.com/fastgeert/go-lfmap
```

## Usage

```go
myMap := NewLFmap()
myMap.Set("key1", "value1")
if tmpValueAsInterface, exists := myMap.Get("key1"); exists {
    if valueAsString, ok := tmpValueAsInterface.(string); ok {
        fmt.Fprintf(os.Stdout, "key1: '%s'", valueAsString)
    }
}
myMap.Exists("key1")
myMap.Remove("key1")
```

## License

Apache 2.0