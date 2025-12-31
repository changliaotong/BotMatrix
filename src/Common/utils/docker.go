package utils

import (
	"log"

	"github.com/docker/docker/client"
)

// InitDockerClient initializes a Docker client
func InitDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Failed to initialize Docker client: %v", err)
		return nil, err
	}
	log.Println("Docker client initialized successfully")
	return cli, nil
}
