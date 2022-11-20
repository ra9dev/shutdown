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
		Addr:    ":8090",
		Handler: mux,
	}

	gracefulShutdownDone := shutdown.Wait()

	shutdown.MustAdd("http_server", func(ctx context.Context) {
		log.Println("started http_server shutdown")

		if err := httpSrv.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown http_server: %v", err)

			return
		}

		log.Println("finished http_server shutdown")
	})

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("failed to listen&serve http_server: %v", err)

			return
		}
	}()

	if err := <-gracefulShutdownDone; err != nil {
		log.Printf("failed to shutdown: %v", err)

		return
	}

	log.Println("successfully shutdown")
}
