Noise
=====

Go client implemention for github.com/eleme/noise.

Install
-------

    go get github.com/eleme/noise/clients/go

Usage
-----

```go
import (
    "fmt"
    "github.com/eleme/noise/clients/go/noise"
)

func main() {
    client := noise.NewNoiseClient("0.0.0.0", 9000)
    client.Sub(func(name string, stamp int, value float64, anoma float64) {
        fmt.Printf("%s %d %.3f %.3f\n", name, stamp, value, anoma)
    })
}
```
