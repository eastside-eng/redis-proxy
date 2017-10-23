package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/eastside-eng/redis-proxy/cache"
	"github.com/eastside-eng/redis-proxy/proxy"
	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var redisAddr string
var redisPassword string
var redisDb int

var cacheTTLMs int
var cachePeriodMs int
var cacheCapacity int

var port int

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "redis-proxy",
	Short: "A simple in-memory Redis proxy. Supports RESP.",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		client := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDb,
		})

		pong, err := client.Ping().Result()
		fmt.Println(pong, err)

		cache, err := cache.NewDecayingLRUCache(cacheCapacity,
			time.Duration(cachePeriodMs)*time.Millisecond,
			time.Duration(cacheTTLMs)*time.Millisecond)

		if err != nil {
			panic("Error creating LRU!")
		}

		server := proxy.NewServer(cache, client)
		server.Run(port)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

	RootCmd.Flags().StringVar(&redisAddr, "redis-hostname", "localhost:6379", "The hostname for the backing redis cache.")
	RootCmd.Flags().StringVar(&redisPassword, "redis-password", "", "The password for the backing redis cache.")
	RootCmd.Flags().IntVar(&redisDb, "redis-database", 0, "The redis database to use. See https://redis.io/commands/select.")

	RootCmd.Flags().IntVar(&cacheCapacity, "capacity", 1024, "The maximum number of entries to cache.")
	RootCmd.Flags().IntVar(&cachePeriodMs, "cache-period", 100, "The periodicity of the cache eviction thread, in milliseconds.")
	RootCmd.Flags().IntVar(&cacheTTLMs, "cache-ttl", 5*60*1000, "A global TTL for cache entries, in milliseconds.")

	RootCmd.Flags().IntVarP(&port, "port", "p", 8001, "A open port used for listening.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
