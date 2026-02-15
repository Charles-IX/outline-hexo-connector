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

	if *isTestMode {
		http.HandleFunc("/webhook", test.PrintWebhook)
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
		http.HandleFunc("/webhook", outlineClient.HandleWebhook)
	}

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
