package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/vulpemventures/nigiri/cli/constants"
)

// Env implements functions for interacting with the environment file used by
// docker-compose to dynamically set service ports and variables
type Env struct{}

func (e *Env) writeEnvForCompose(path, strJSON string) error {
	var env map[string]interface{}
	err := json.Unmarshal([]byte(strJSON), &env)
	if err != nil {
		return constants.ErrMalformedJSON
	}

	fileContent := ""
	for chain, services := range env["ports"].(map[string]interface{}) {
		for k, v := range services.(map[string]interface{}) {
			fileContent += fmt.Sprintf("%s_%s_PORT=%d\n", strings.ToUpper(chain), strings.ToUpper(k), int(v.(float64)))
		}
	}
	for hostname, url := range env["urls"].(map[string]interface{}) {
		fileContent += fmt.Sprintf("%s_URL=%s\n", strings.ToUpper(hostname), url.(string))
	}

	return ioutil.WriteFile(path, []byte(fileContent), os.ModePerm)
}

func (e *Env) readEnvForCompose(path string) (map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	ports := map[string]map[string]int{
		"bitcoin": map[string]int{},
		"liquid":  map[string]int{},
	}
	urls := map[string]string{}
	// Each line is in the format PREFIX_SERVICE_NAME_SUFFIX=value
	// PREFIX is either 'BITCOIN' or 'LIQUID', while SUFFIX is either 'PORT' or 'URL'
	for scanner.Scan() {
		line := scanner.Text()
		splitLine := strings.Split(line, "=")
		key := splitLine[0]

		if strings.Contains(key, "PORT") {
			value, _ := strconv.Atoi(splitLine[1])
			chain := "bitcoin"
			if strings.HasPrefix(key, strings.ToUpper("liquid")) {
				chain = "liquid"
			}
			prefix := strings.ToUpper(fmt.Sprintf("%s_", chain))
			suffix := "_PORT"
			trimmedKey := strings.ToLower(
				strings.TrimSuffix(strings.TrimPrefix(key, prefix), suffix),
			)
			ports[chain][trimmedKey] = value
		} else {
			// Here the prefix is not trimmed
			value := splitLine[1]
			suffix := "_URL"
			trimmedKey := strings.ToLower(strings.TrimSuffix(key, suffix))
			urls[trimmedKey] = value
		}
	}

	return map[string]interface{}{"ports": ports, "urls": urls}, nil
}

func (e *Env) load(path string) []string {
	content, _ := ioutil.ReadFile(path)
	lines := strings.Split(string(content), "\n")
	env := os.Environ()
	for _, line := range lines {
		if line != "" {
			env = append(env, line)
		}
	}

	return env
}

type envPortsData struct {
	Node       int `json:"node,omitempty"`
	Esplora    int `json:"esplora,omitempty"`
	Electrs    int `json:"electrs,omitempty"`
	ElectrsRPC int `json:"electrs_rpc,omitempty"`
	Chopsticks int `json:"chopsticks,omitempty"`
}
type envPorts struct {
	Bitcoin *envPortsData `json:"bitcoin,omitempty"`
	Liquid  *envPortsData `json:"liquid,omitempty"`
}
type envUrls struct {
	BitcoinEsplora string `json:"bitcoin_esplora,omitempty"`
	LiquidEsplora  string `json:"liquid_esplora,omitempty"`
}
type envJSON struct {
	Ports *envPorts `json:"ports,omitempty"`
	Urls  *envUrls  `json:"urls,omitempty"`
}

func (e envJSON) copy() envJSON {
	var v envJSON
	bytes, _ := json.Marshal(e)
	json.Unmarshal(bytes, &v)

	return v
}
