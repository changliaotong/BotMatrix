package sandbox

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	clog "BotMatrix/common/log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type SandboxManager struct {
	cli          *client.Client
	defaultImage string
}

type Sandbox struct {
	ID      string // Container ID
	Manager *SandboxManager
}

func NewSandboxManager(cli *client.Client, defaultImage string) *SandboxManager {
	if defaultImage == "" {
		defaultImage = "python:3.10-slim" // Default to Python environment for versatility
	}
	return &SandboxManager{
		cli:          cli,
		defaultImage: defaultImage,
	}
}

// CreateSandbox starts a new isolated container
func (m *SandboxManager) CreateSandbox(ctx context.Context, image string) (*Sandbox, error) {
	if image == "" {
		image = m.defaultImage
	}

	// Ensure image exists
	// Try to pull the image first. We ignore errors here because the image might already exist locally
	// or the network might be down, in which case we rely on the local cache.
	reader, err := m.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err == nil {
		io.Copy(io.Discard, reader)
		reader.Close()
	} else {
		clog.Warn(fmt.Sprintf("[Sandbox] Failed to pull image %s: %v. Trying to use local version.", image, err))
	}

	resp, err := m.cli.ContainerCreate(ctx, &container.Config{
		Image:        image,
		Cmd:          []string{"tail", "-f", "/dev/null"}, // Keep running
		Tty:          false,
		OpenStdin:    false,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
	}, &container.HostConfig{
		AutoRemove: true, // Clean up on exit
		Resources: container.Resources{
			Memory:   512 * 1024 * 1024, // 512MB limit
			NanoCPUs: 500000000,         // 0.5 CPU
		},
	}, nil, nil, "")

	if err != nil {
		clog.Error("[Sandbox] Failed to create container: " + err.Error())
		return nil, err
	}

	if err := m.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		clog.Error("[Sandbox] Failed to start container: " + err.Error())
		return nil, err
	}

	clog.Info(fmt.Sprintf("[Sandbox] Created sandbox container %s", resp.ID[:12]))

	// Initialize workspace
	sandbox := &Sandbox{ID: resp.ID, Manager: m}
	sandbox.Exec(ctx, "mkdir -p /workspace")

	return sandbox, nil
}

// Exec executes a command in the sandbox and returns stdout, stderr
func (s *Sandbox) Exec(ctx context.Context, cmd string) (string, string, error) {
	// Create Exec
	execConfig := types.ExecConfig{
		Cmd:          []string{"/bin/sh", "-c", cmd},
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := s.Manager.cli.ContainerExecCreate(ctx, s.ID, execConfig)
	if err != nil {
		return "", "", err
	}

	// Attach
	resp, err := s.Manager.cli.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", "", err
	}
	defer resp.Close()

	// Capture output
	var stdoutBuf, stderrBuf bytes.Buffer
	// Docker multiplexes stdout and stderr, stdcopy splits them
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, resp.Reader)
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), nil
}

// WriteFile writes content to a file in the sandbox
func (s *Sandbox) WriteFile(ctx context.Context, path string, content []byte) error {
	// Create a tar archive containing the file
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: path,
		Mode: 0644,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	// Upload to container
	// Note: Path should be directory, but CopyToContainer treats path as destination directory
	// If we want to write to /workspace/file.txt, we copy to /workspace
	// The tar header name must be relative or match what we want

	// Simplification: We assume path is relative to /workspace or absolute.
	// If absolute, we need to handle the parent dir.
	// For simplicity, let's treat the tar header name as the full path (or relative to CopyToContainer path)
	// Actually, CopyToContainer extracts the tar at the destination path.

	return s.Manager.cli.CopyToContainer(ctx, s.ID, "/", &buf, types.CopyToContainerOptions{})
}

// ReadFile reads content from a file in the sandbox
func (s *Sandbox) ReadFile(ctx context.Context, path string) ([]byte, error) {
	reader, _, err := s.Manager.cli.CopyFromContainer(ctx, s.ID, path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	tr := tar.NewReader(reader)

	// We expect the first file to be the one we asked for
	_, err = tr.Next()
	if err != nil {
		return nil, err
	}

	return io.ReadAll(tr)
}

// Destroy stops and removes the sandbox
func (s *Sandbox) Destroy(ctx context.Context) error {
	// Use a timeout for stopping
	timeout := 5 // seconds
	stopOptions := container.StopOptions{Timeout: &timeout}
	return s.Manager.cli.ContainerStop(ctx, s.ID, stopOptions)
}
