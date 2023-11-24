package task

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

// Configuration for Docker container
type Config struct {
	Name          string // To Idenity a task in the Orch Sys
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string // Image name that container runs
	Cpu           float64
	Memory        int64    // To tell docker-daemon about the memory required for a task
	Disk          int64    // To tell docker-daemon about the space required for a task
	Env           []string // Specify environment variables that passed into the container
	RestartPolicy string   // Specify to docker-daemon what to do when the container fails
}

// Docker API configurations with Docker Client attached
type Docker struct {
	Client      *client.Client
	Config      Config
	ContainerId string
}

// Return value of Docker instance
type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

// Run Container
func (d* Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})
	if err != nil{
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
	}

	io.Copy(os.Stdout, reader)

	rp := container.RestartPolicy{Name: d.Config.RestartPolicy}

	r := container.Resources{Memory: d.Config.Memory}

	cc := container.Config{Image: d.Config.Image, Env: d.Config.Env}

	hc := container.HostConfig{RestartPolicy: rp, Resources: r, PublishAllPorts: true}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil{
		log.Printf("Error creaing container using image %s: %v\n", d.Config.Image, err)

		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil{
		log.Printf("Error starting container using image %s: %v\n", resp.ID, err)

		return DockerResult{Error: err}
	}

	d.ContainerId = resp.ID

	out, err := d.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil{
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)

		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{
		ContainerId: resp.ID,
		Action: "Start",
		Result: "success",
	}
}