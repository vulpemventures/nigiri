package controller

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/vulpemventures/nigiri/cli/constants"
)

// Docker type handles interfaction with containers via docker and docker-compose
type Docker struct {
	cli *client.Client
}

// New initialize a new Docker handler
func (d *Docker) New() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	d.cli = cli
	return nil
}

func (d *Docker) isDockerRunning() bool {
	_, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return false
	}
	return true
}

func (d *Docker) findNigiriContainers(listAllContainers bool) bool {
	containers, _ := d.cli.ContainerList(context.Background(), types.ContainerListOptions{All: listAllContainers})

	if len(containers) <= 0 {
		return false
	}

	images := []string{}
	for _, c := range containers {
		images = append(images, c.Image)
	}

	for _, nigiriImage := range constants.NigiriImages {
		// just check if services for bitcoin chain are up and running
		if !strings.Contains(nigiriImage, "liquid") {
			if !contains(images, nigiriImage) {
				return false
			}
		}
	}
	return true
}

func (d *Docker) isNigiriRunning() bool {
	return d.findNigiriContainers(false)
}

func (d *Docker) isNigiriStopped() bool {
	isRunning := d.isNigiriRunning()
	if !isRunning {
		return d.findNigiriContainers(true)
	}
	return false
}

func (d *Docker) cleanVolumes(path string) error {
	subdirs, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, d := range subdirs {
		path := filepath.Join(path, d.Name())
		subsubdirs, _ := ioutil.ReadDir(path)
		for _, sd := range subsubdirs {
			if sd.IsDir() {
				if err := os.RemoveAll(filepath.Join(path, sd.Name())); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func contains(list []string, elem string) bool {
	for _, l := range list {
		if l == elem {
			return true
		}
	}
	return false
}
