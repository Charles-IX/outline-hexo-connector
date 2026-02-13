package main

import (
	"log"
	"net/http"
	"outline-hexo-connector/internal/config"
	"outline-hexo-connector/internal/outline"
	"outline-hexo-connector/internal/test"

	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	port := flag.StringP("port", "p", "9000", "Port to listen on for webhook requests")
	isTestMode := flag.BoolP("test", "t", false, "Run in test mode to print raw incoming requests")
	configFile := flag.StringP("config", "c", "config.yaml", "Path to config file")
	flag.Parse()

	if *isTestMode {
		http.HandleFunc("/webhook", test.PrintWebhook)
		log.Println("Running in test mode - Print raw incoming requests only")
	} else {
		cfg, err := config.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading config - %v", err)
		}
		log.Printf("Config loaded from %s", *configFile)

		outlineClient := outline.NewClient(cfg)
		http.HandleFunc("/webhook", outlineClient.HandleWebhook)
	}

	log.Printf("Webhook listener started at port %s...", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Error starting server - %v", err)
	}
}
