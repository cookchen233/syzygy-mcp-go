package main

import (
	"log"
	"os"

	"github.com/cookchen233/syzygy-mcp-go/internal/interface/mcp"
)

func main() {
	logger := log.New(os.Stderr, "syzygy-mcp: ", log.LstdFlags|log.LUTC)

	srv := mcp.NewServer(mcp.ServerConfig{
		Name:    "syzygy-mcp",
		Version: "0.1.0",
		Logger:  logger,
	})

	if err := srv.Run(); err != nil {
		logger.Printf("server stopped with error: %v", err)
		os.Exit(1)
	}
}
