package common

import (
	"log"

	"github.com/docker/docker/client"
)

// InitDockerClient initializes the Docker client for the manager
func (m *Manager) InitDockerClient() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Failed to initialize Docker client: %v", err)
		return err
	}
	m.DockerClient = cli
	log.Println("Docker client initialized successfully")
	return nil
}
