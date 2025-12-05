package main

import (
	"os"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/server"
	"github.com/adriein/tibia-char/pkg/constants"
	_ "github.com/lib/pq"
)

func main() {
	internal.NewApp()

	server.New(os.Getenv(constants.ServerPort))
}
