package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sa06/prometheus-pushgateway/src/metrics"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		signal.Stop(c)
		cancel()
	}()

	gw := os.Getenv("APP_PUSH_GATEWAY")
	if gw == "" {
		gw = "http://localhost:9091"
	}

	instance := os.Getenv("APP_INSTANCE")

	cfg := &metrics.PushConfig{
		Instance: instance,
		Job:      "app-service",
		Interval: 5 * time.Second,
		URL:      gw,
	}

	// init push metrics
	metrics.InitPusher(ctx, cfg)

	runJob(ctx, 5*time.Second)

	//
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})
	mux.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Printf("Listening port: %s \n", httpServer.Addr)
		log.Println(httpServer.ListenAndServe())
	}()

	select {
	case <-c:
		cancel()
		log.Println("cancel")
	case <-ctx.Done():
		log.Println("done")
	}

	_ = httpServer.Shutdown(ctx)
}

func runJob(ctx context.Context, interval time.Duration) {
	service := func() int {
		n := 1 + rand.Intn(7)
		time.Sleep(time.Duration(n) * time.Millisecond)
		return n
	}

	name := "action-1"

	// chain metrics middleware
	service = metrics.ServiceMiddleware(name, service)

	go func(name string, action func() int) {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				service()
				log.Println(name)
			case <-ctx.Done():
				return
			}
		}
	}(name, service)

}
