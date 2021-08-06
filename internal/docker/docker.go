package docker

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/compose-spec/compose-go/loader"
)

func GetServices(composeFile string) ([][]string, error) {

	composeBytes, err := ioutil.ReadFile(composeFile)
	if err != nil {
		return nil, err
	}

	parsed, err := loader.ParseYAML(composeBytes)
	if err != nil {
		return nil, err
	}

	if _, ok := parsed["services"]; !ok {
		return nil, errors.New("missing services in compose")
	}

	serviceMap := parsed["services"].(map[string]interface{})

	var services [][]string
	for k, v := range serviceMap {
		m := v.(map[string]interface{})
		i := m["ports"].([]interface{})
		for _, j := range i {
			port := j.(string)
			exposedPorts := strings.Split(port, ":")
			endpoint := "localhost:" + exposedPorts[0]
			services = append(services, []string{k, endpoint})
		}

	}

	return services, nil
}

func GetPortsForService(composeFile string, serviceName string) ([]string, error) {

	composeBytes, err := ioutil.ReadFile(composeFile)
	if err != nil {
		return nil, err
	}

	parsed, err := loader.ParseYAML(composeBytes)
	if err != nil {
		return nil, err
	}

	if _, ok := parsed["services"]; !ok {
		return nil, errors.New("missing services in compose")
	}

	serviceMap := parsed["services"].(map[string]interface{})

	var ports []string
	for k, v := range serviceMap {
		if k == serviceName {
			m := v.(map[string]interface{})
			i := m["ports"].([]interface{})
			for _, j := range i {
				port := j.(string)
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}
