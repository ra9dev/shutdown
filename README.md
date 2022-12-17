# shutdown

Graceful shutdown for Go! It listens process termination signals and handles
your shutdown callbacks!

## Features

- Scopes: local, global
- Dependant callbacks with concurrent execution, e.g: db -> [http_server, grpc_server] -> cache

## Example:

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/ra9dev/shutdown"
)

func main() {
	mux := http.NewServeMux()

	httpSrv := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	shutdown.MustAdd("http_server", func(ctx context.Context) {
		if err := httpSrv.Shutdown(ctx); err != nil {
			log.Println("failed to shut down http server")

			return
		}

		log.Println("gracefully shut down http server")
	})

	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	if err := shutdown.Wait(); err != nil {
		panic(err)
	}
}
```