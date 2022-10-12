package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/Propertyfinder/shutdown"
)

func main() {
	mux := http.NewServeMux()

	httpSrv := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	shutdown.Add(func(ctx context.Context) {
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
