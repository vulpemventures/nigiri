package docker

import (
	"errors"
	"io/ioutil"

	"github.com/compose-spec/compose-go/loader"
)

func GetServices(composeFile string) ([]string, error) {

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

	var services []string
	for k := range serviceMap {
		services = append(services, k)
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
