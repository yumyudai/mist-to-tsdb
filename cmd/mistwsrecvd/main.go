package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"mist-to-tsdb/internal/mistwsrcvr"
)

func main() {
	var err error
	var configFile string
	var config mistwsrcvr.Config

	rootCmd := &cobra.Command {
		Use: "mistwsrecvd",
		Short: "Receive data from Mist WebSocket API and write to TSDB",
		// Main Entry Point
		Run: func(c *cobra.Command, args []string) {
			// Init 
			rcvr, err := mistwsrcvr.New(config)
			if err != nil {
				log.Fatalf("Failed on init: %v", err)
			}

			err = rcvr.Run()
			if err != nil {
				log.Fatalf("Failed on start: %v", err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.json", "Path to configuration")

	// Default Values
	viper.SetDefault("mist.endpoint", "api-ws.mist.com")
	viper.SetDefault("tsdb.enabled", false)
	viper.SetDefault("tsdb.debug", false)
	viper.SetDefault("tsdb.driver", "awstimestream")
	viper.SetDefault("tsdb.bufsize", 128)
	viper.SetDefault("tsdb.awstimestream.region", "us-east-1")
	viper.SetDefault("tsdb.awstimestream.maxretries", 3)
	viper.SetDefault("pubsub.enabled", false)
	viper.SetDefault("pubsub.debug", false)
	viper.SetDefault("pubsub.driver", "kafka")
	viper.SetDefault("pubsub.kafka.async", true)
	viper.SetDefault("pubsub.bufsize", 128)

	// Read Configuration File Before Start
	cobra.OnInitialize(func() {
		_, err := os.Stat(configFile)
		if os.IsNotExist(err) {
			log.Fatalf("Config file %s does not exist!", configFile)
		}

		viper.SetConfigFile(configFile)
		viper.SetConfigType("json")
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatalf("Failed to read config: %v", err)
		}

		err = viper.Unmarshal(&config)
		if err != nil {
			log.Fatalf("Failed to parse config: %v", err)
		}

		log.Printf("Loaded config file: %s", configFile)
	})

	// Launch (cobra.OnInitializa -> rootCmd.Run)
	err = rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

}
