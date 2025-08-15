package main

import (
	"fmt"
	"go-certdist/command/client"
	"go-certdist/command/server"
	"go-certdist/common"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	common.InitLogger()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		if len(os.Args) < 3 {
			log.Fatal().Msg("Usage: go-certdist server <config file path>")
		}
		config := common.LoadServerConfig(os.Args[2])
		server.StartServer(config)
	case "client":
		if len(os.Args) < 3 {
			log.Fatal().Msg("Usage: go-certdist client <config file path>")
		}
		config := common.LoadClientConfig(os.Args[2])
		client.ExecuteClient(config)
	case "keygen":
		if err := common.GenerateAndPrintKeyPair(); err != nil {
			log.Fatal().Err(err).Msg("Failed to generate key pair")
		}
	case "config":
		if len(os.Args) < 3 {
			log.Fatal().Msg("Usage: go-certdist config <server|client>")
		}
		switch os.Args[2] {
		case "server":
			common.WriteDummyServerConfig()
		case "client":
			common.WriteDummyClientConfig()
		default:
			log.Fatal().Str("command", os.Args[2]).Msg("Unknown config command")
		}
	case "help":
		printHelp()
	default:
		printHelp()
		log.Fatal().Str("command", os.Args[1]).Msg("Unknown command")
	}
}

func printHelp() {
	fmt.Println(`go-certdist - A tool for distributing certificates.

Usage:

	go-certdist <command> [arguments]

The commands are:

	server <path>      start the server
	client <path>      start the client
	keygen             generate a new age key pair
	config <type>      write a dummy config file (server or client)
	help               print this help message
	`)
}
