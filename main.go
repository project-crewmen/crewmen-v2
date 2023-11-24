package main

import(
	"crewmen-v2/task"
	"github.com/docker/docker/client"
)

func main()  {
	c := task.Config{
		Name: "test",
		Image: "strm/helloworld-http",
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)

	d := task.Docker{
		Client: dc,
		Config: c,
	}

	d.Run()
}