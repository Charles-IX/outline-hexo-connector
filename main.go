package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"outline-hexo-connector/internal/config"
	"outline-hexo-connector/internal/hexo"
	"outline-hexo-connector/internal/outline"
	"outline-hexo-connector/internal/server"
	"outline-hexo-connector/internal/test"
	"syscall"

	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	port := flag.StringP("port", "p", "9000", "Port to listen on for webhook requests")
	isTestMode := flag.BoolP("test", "t", false, "Run in test mode to print raw incoming requests")
	configFile := flag.StringP("config", "c", "config.yaml", "Path to config file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var webhookPostHandler http.HandlerFunc
	var directAccessHandler http.Handler

	if *isTestMode {
		webhookPostHandler = test.PrintWebhook
		handler, err := server.NewDirectAccessHandler(nil)
		if err != nil {
			log.Fatalf("Error creating direct access handler - %v", err)
		}
		directAccessHandler = handler
		log.Printf("Running in test mode - Print raw incoming requests only")
	} else {
		cfg, err := config.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading config - %v", err)
		}
		log.Printf("Config loaded from %s", *configFile)

		hexoTrigger := hexo.NewTrigger(cfg)
		hexoTrigger.Watch(ctx)
		outlineClient := outline.NewClient(cfg, hexoTrigger)
		webhookPostHandler = outlineClient.HandleWebhook
		handler, err := server.NewDirectAccessHandler(cfg)
		if err != nil {
			log.Fatalf("Error creating direct access handler - %v", err)
		}
		directAccessHandler = handler
	}

	http.Handle("/", directAccessHandler)
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			webhookPostHandler(w, r)
		case http.MethodGet, http.MethodHead:
			directAccessHandler.ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "POST, GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	go func() {
		err := http.ListenAndServe(":"+*port, nil)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server - %v", err)
		}
	}()
	log.Printf("Webhook listener started at port %s", *port)

	<-ctx.Done()
	log.Printf("Stop listening for Outline webhook requests")
}
