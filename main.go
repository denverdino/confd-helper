package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/denverdino/shlex"
	"github.com/samalba/dockerclient"
	"os"
	"path"
)

func ExecInContainer(client *dockerclient.DockerClient, containerId string, cmd []string) (string, error) {
	// If native exec support does not exist in the local docker daemon use nsinit.

	execConfig := &dockerclient.ExecConfig{
		Container:    containerId,
		Cmd:          cmd,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	return client.Exec(execConfig)
}

var (
	flAddr = cli.StringFlag{
		Name:  "addr",
		Value: "unix:///var/run/docker.sock",
		Usage: "URI to Docker daemon",
	}

	flContainer = cli.StringFlag{
		Name:  "container",
		Value: "",
		Usage: "Container Id or name",
	}

	flCommand = cli.StringFlag{
		Name:  "command, c",
		Value: "",
		Usage: "Command for execution",
	}
	flSignal = cli.StringFlag{
		Name:  "signal, s",
		Value: "",
		Usage: "Signal to send to the container",
	}
)

func main() {
	app := cli.NewApp()

	app.Name = path.Base(os.Args[0])
	app.Usage = "Confd helper for Docker"
	app.Version = "0.1.0"
	app.Author = ""
	app.Email = ""

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug mode",
			EnvVar: "DEBUG",
		},

		cli.StringFlag{
			Name:  "log-level, l",
			Value: "info",
			Usage: fmt.Sprintf("Log level (options: debug, info, warn, error, fatal, panic)"),
		},
	}

	// logs
	app.Before = func(c *cli.Context) error {
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.SetLevel(level)

		// If a log level wasn't specified and we are running in debug mode,
		// enforce log-level=debug.
		if !c.IsSet("log-level") && !c.IsSet("l") && c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:      "exec",
			ShortName: "e",
			Usage:     "Run a command in a running container",
			Flags: []cli.Flag{
				flAddr,
				flContainer,
				flCommand,
			},
			Action: exec,
		},
		{
			Name:      "kill",
			ShortName: "k",
			Usage:     "Kill a running container using SIGKILL or a specified signal",
			Flags: []cli.Flag{
				flAddr,
				flContainer,
				flSignal,
			},
			Action: kill,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getDockerClient(c *cli.Context) *dockerclient.DockerClient {
	path := c.String("addr")

	if path == "" {
		log.Fatalf("addr is required to connect Docker Daemon. See '%s --help'.", c.App.Name)
	}

	dockerClient, err := dockerclient.NewDockerClient(path, nil)

	if err != nil {
		log.Fatalf("Failed to get Docker Client: %v", err)
	}

	return dockerClient
}

func exec(c *cli.Context) {
	client := getDockerClient(c)
	container := c.String("container")
	cmd := c.String("command")
	log.Infof("docker exec %s %s", container, cmd)
	cmds, err := shlex.Split(cmd)

	if err != nil {
		log.Fatalf("Failed to split command '%s': %v", cmd, err)
	}

	execConfig := &dockerclient.ExecConfig{
		Container:    container,
		Detach:       false,
		Cmd:          cmds,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	result, err := client.Exec(execConfig)
	if err != nil {
		log.Fatalf("Failed to execute command '%s' in container '%s': %v", cmd, container, err)
	} else {
		log.Infof("Exec id: %s", result)
	}

}

func kill(c *cli.Context) {
	client := getDockerClient(c)
	container := c.String("container")
	signal := c.String("signal")
	err := client.KillContainer(container, signal)
	if err != nil {
		log.Fatalf("Failed to kill container '%s' with signal '%s': %v", container, signal, err)
	}
}
