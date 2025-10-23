package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	grpcapi "github.com/treerootboy/anyalert/api/grpc"
	httpapi "github.com/treerootboy/anyalert/api/http"
	"github.com/treerootboy/anyalert/internal/config"
	"github.com/treerootboy/anyalert/pkg/notifier"
	"github.com/treerootboy/anyalert/pkg/plugins/phone"
	"github.com/treerootboy/anyalert/pkg/plugins/slack"
	"github.com/treerootboy/anyalert/pkg/plugins/sms"
	"github.com/treerootboy/anyalert/pkg/plugins/youdu"
)

func main() {
	configFile := flag.String("config", "config.json", "Configuration file path")
	flag.Parse()

	// Load configuration
	cfg, err := loadConfiguration(*configFile)
	if err != nil {
		log.Printf("Failed to load config file, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}

	// Create notification manager
	manager := notifier.NewManager()

	// Register notification channels
	if err := registerChannels(manager, cfg); err != nil {
		log.Fatalf("Failed to register channels: %v", err)
	}

	log.Printf("Registered channels: %v", manager.ListChannels())

	// Start servers
	errChan := make(chan error, 2)

	if cfg.Server.HTTP.Enabled {
		go func() {
			httpServer := httpapi.NewServer(manager, cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
			errChan <- httpServer.Start()
		}()
	}

	if cfg.Server.GRPC.Enabled {
		go func() {
			grpcServer := grpcapi.NewServer(manager, cfg.Server.GRPC.Host, cfg.Server.GRPC.Port)
			errChan <- grpcServer.Start()
		}()
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	case <-sigChan:
		log.Println("Shutting down...")
	}
}

func loadConfiguration(filename string) (*config.Config, error) {
	return config.LoadConfig(filename)
}

func registerChannels(manager *notifier.Manager, cfg *config.Config) error {
	for name, channelCfg := range cfg.Channels {
		if !channelCfg.Enabled {
			continue
		}

		var n notifier.Notifier
		switch channelCfg.Type {
		case "slack":
			n = slack.NewSlack()
		case "sms":
			n = sms.NewSMS()
		case "phone":
			n = phone.NewPhone()
		case "youdu":
			n = youdu.NewYoudu()
		default:
			log.Printf("Unknown channel type: %s", channelCfg.Type)
			continue
		}

		if err := n.Initialize(channelCfg.Config); err != nil {
			log.Printf("Failed to initialize %s: %v", name, err)
			continue
		}

		if err := manager.Register(n); err != nil {
			log.Printf("Failed to register %s: %v", name, err)
			continue
		}

		log.Printf("Registered channel: %s (type: %s)", name, channelCfg.Type)
	}

	return nil
}
