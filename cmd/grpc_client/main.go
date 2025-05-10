package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/KaffeeMaschina/subpub/config"
	"github.com/KaffeeMaschina/subpub/config/env"
	pubsubclient "github.com/KaffeeMaschina/subpub/pkg/PubSubClient"
	"log"
	"time"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}
func main() {
	flag.Parse()
	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	grpcConfig, err := env.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	c := pubsubclient.NewClient(grpcConfig.Address())

	err = c.Publish(context.Background(), "news", "Привет, мир!")
	if err != nil {
		log.Printf("publish failed: %v", err)
	}
	log.Println("Message published")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = c.Subscribe(ctx, "news", func(data string) {
		fmt.Println(data)
	})
	if err != nil {
		log.Printf("subscribe failed: %v", err)
	}

	time.Sleep(2 * time.Second)
	err = c.Publish(context.Background(), "news", "Второе сообщение")
	if err != nil {
		log.Printf("publish failed: %v", err)
	}

	time.Sleep(3 * time.Second)
}
