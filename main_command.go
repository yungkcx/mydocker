package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/yungkcx/mydocker/cgroups/subsystems"
	"github.com/yungkcx/mydocker/container"
	"github.com/yungkcx/mydocker/image"
)

var runCommand = cli.Command{
	Name:      "run",
	ArgsUsage: "IMAGE",
	Usage:     `Create a container with namespace and cgroups limit mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
	},
	Action: func(context *cli.Context) error {
		imageName := context.Args().Get(0)
		tty := context.Bool("ti")
		volume := context.String("v")
		cmdArray := []string{}
		if context.NArg() < 1 {
			return fmt.Errorf("Missing image name")
		} else if context.NArg() == 1 {
			cmdArray = append(cmdArray, "sh")
		} else {
			for _, arg := range context.Args()[1:] {
				cmdArray = append(cmdArray, arg)
			}
		}
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CPUSet:      context.String("cpuset"),
			CPUShare:    context.String("cpushare"),
		}
		return Run(imageName, tty, volume, cmdArray, resConf)
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		// Will never return if success.
		return container.RunContainerInitProcess()
	},
}

var imagesCommand = cli.Command{
	Name:  "images",
	Usage: "List images",
	Action: func(context *cli.Context) error {
		return image.ListImages()
	},
}

var imageCommand = cli.Command{
	Name:      "image",
	ArgsUsage: "FILE",
	Usage:     "Create an image using from tar file",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "o",
			Usage: "image's `name`",
		},
	},
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing tar file")
		}
		name := context.String("o")
		return image.CreateImage(context.Args().Get(0), name)
	},
}

var psCommand = cli.Command{
	Name:  "ps",
	Usage: "List containers",
	Action: func(context *cli.Context) error {
		return container.ListContainers()
	},
}

var rmiCommand = cli.Command{
	Name:      "rmi",
	ArgsUsage: "IMAGE...",
	Usage:     "Remove image",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing image name")
		}
		names := context.Args()
		return image.RemoveImage(names...)
	},
}
