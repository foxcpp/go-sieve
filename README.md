go-sieve
====================

Sieve email filtering language ([RFC 5228][rfc5228])
implementation in Go.

## Extensions

- envelope (RFC 5228)
- fileinto (RFC 5228)

## Example

```go
package main

import (
	"context"
	"strings"

	"github.com/foxcpp/go-sieve"
	"github.com/foxcpp/go-sieve/interp"
)

func main() {
	const script = `
require "fileinto";
if header :contains "subject" "rich" {
    fileinto "Junk";
}
`

	parsed, _ := sieve.Load(strings.NewReader(script), sieve.DefaultOptions())
	data := interp.NewRuntimeData(
		parsed,
		interp.Callback{},
	)
	parsed.Execute(context.Background(), data)
	// inspect RuntimeData for results
}

```

[rfc5228]: https://datatracker.ietf.org/doc/html/rfc5228