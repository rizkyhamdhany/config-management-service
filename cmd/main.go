package main

import (
	"configuration-management-service/pkg/app"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	srv, shutdown, err := app.BuildServer()
	if err != nil {
		log.Fatalf("boot: %v", err)
	}

	go func() {
		if err := srv.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
