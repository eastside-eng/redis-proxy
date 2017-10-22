package main

import (
	"github.com/eastside-eng/redis-proxy/cmd"
	log "github.com/eastside-eng/redis-proxy/internal/log"
)

func main() {
	// Wiring up logging prior to invoking any subcommands.
	logger := log.NewLogger()
	// We must defer Sync to ensure that all logs are output prior to exiting.
	defer logger.Sync()
	log.SetLogger(logger)

	// The Cobra command entry point. Will parse args and find a matching
	// command handler.
	cmd.Execute()
}
