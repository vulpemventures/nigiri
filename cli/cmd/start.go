package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/altafan/nigiri/cli/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

const listAll = true

var StartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Build and start Nigiri",
	RunE:    start,
	PreRunE: startChecks,
}

func startChecks(cmd *cobra.Command, args []string) error {
	network, _ := cmd.Flags().GetString("network")
	datadir, _ := cmd.Flags().GetString("datadir")
	ports, _ := cmd.Flags().GetString("ports")

	// check flags
	if !isNetworkOk(network) {
		return fmt.Errorf("Invalid network: %s", network)
	}

	if !isDatadirOk(datadir) {
		return fmt.Errorf("Invalid datadir, it must be an absolute path: %s", datadir)
	}
	if !isEnvOk(ports) {
		return fmt.Errorf("Invalid env JSON, it must contain a \"bitcoin\" object with at least one service specified. It can optionally contain a \"liquid\" object with at least one service specified.\nGot: %s", ports)
	}

	// if nigiri is already running return error
	isRunning, err := nigiriIsRunning()
	if err != nil {
		return err
	}
	if isRunning {
		return fmt.Errorf("Nigiri is already running, please stop it first")
	}

	// scratch datadir if not exists
	if err := os.MkdirAll(datadir, 0755); err != nil {
		return err
	}

	// if datadir is set we must copy the resources directory from ~/.nigiri
	// to the new one
	if datadir != getDefaultDir() {
		if err := copyResources(datadir); err != nil {
			return err
		}
	}

	// if nigiri not exists, we need to write the configuration file and then
	// read from it to get viper updated, otherwise we just read from it.
	exists, err := nigiriExistsAndNotRunning()
	if err != nil {
		return err
	}
	if !exists {
		filedir := getPath(datadir, "config")
		if err := config.WriteConfig(filedir); err != nil {
			return err
		}
		// .env must be in the directory where docker-compose is run from, not where YAML files are placed
		// https://docs.docker.com/compose/env-file/
		filedir = getPath(datadir, "env")
		if err := writeComposeEnvFile(filedir, ports); err != nil {
			return err
		}
	}
	if err := config.ReadFromFile(datadir); err != nil {
		return err
	}

	return nil
}

func start(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	liquidEnabled, _ := cmd.Flags().GetBool("liquid")

	bashCmd, err := getStartBashCmd(datadir)
	if err != nil {
		return err
	}

	if err := bashCmd.Run(); err != nil {
		return err
	}

	path := getPath(datadir, "env")
	ports, err := readComposeEnvFile(path)
	if err != nil {
		return err
	}

	prettyPrintServices := func(chain string, services map[string]int) {
		fmt.Printf("%s services:\n", chain)
		for name, port := range services {
			formatName := fmt.Sprintf("%s:", name)
			fmt.Printf("   %-14s localhost:%d\n", formatName, port)
		}
	}

	for chain, services := range ports {
		if chain == "bitcoin" {
			prettyPrintServices(chain, services)
		} else if liquidEnabled {
			prettyPrintServices(chain, services)
		}
	}

	return nil
}

var images = map[string]bool{
	"vulpemventures/bitcoin:latest":           true,
	"vulpemventures/liquid:latest":            true,
	"vulpemventures/electrs:latest":           true,
	"vulpemventures/electrs-liquid:latest":    true,
	"vulpemventures/esplora:latest":           true,
	"vulpemventures/esplora-liquid:latest":    true,
	"vulpemventures/nigiri-chopsticks:latest": true,
}

func copyResources(datadir string) error {
	defaultDatadir := getDefaultDir()
	cmd := exec.Command("cp", "-R", filepath.Join(defaultDatadir, "resources"), datadir)
	return cmd.Run()

}

func nigiriExists(listAll bool) (bool, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return false, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: listAll})
	if err != nil {
		return false, err
	}

	for _, container := range containers {
		if images[container.Image] {
			return true, nil
		}
	}

	return false, nil
}

func isNetworkOk(network string) bool {
	var ok bool
	for _, n := range []string{"regtest"} {
		if network == n {
			ok = true
		}
	}

	return ok
}

func isDatadirOk(datadir string) bool {
	return filepath.IsAbs(datadir)
}

func isEnvOk(stringifiedJSON string) bool {
	var parsedJSON map[string]map[string]int
	err := json.Unmarshal([]byte(stringifiedJSON), &parsedJSON)
	if err != nil {
		return false
	}

	if len(parsedJSON) <= 0 {
		return false
	}
	if len(parsedJSON["bitcoin"]) <= 0 {
		return false
	}
	if parsedJSON["bitcoin"]["node"] <= 0 &&
		parsedJSON["bitcoin"]["electrs"] <= 0 &&
		parsedJSON["bitcoin"]["esplora"] <= 0 &&
		parsedJSON["bitcoin"]["chopsticks"] <= 0 {
		return false
	}

	if len(parsedJSON["liquid"]) > 0 &&
		parsedJSON["liquid"]["node"] <= 0 &&
		parsedJSON["liquid"]["electrs"] <= 0 &&
		parsedJSON["liquid"]["esplora"] <= 0 &&
		parsedJSON["liquid"]["chopsticks"] <= 0 {
		return false
	}

	return true
}

func getPath(datadir, t string) string {
	viper := config.Viper()

	if t == "compose" {
		network := viper.GetString("network")
		attachLiquid := viper.GetBool("attachLiquid")
		if attachLiquid {
			network += "-liquid"
		}
		return filepath.Join(datadir, "resources", fmt.Sprintf("docker-compose-%s.yml", network))
	}

	if t == "env" {
		return filepath.Join(datadir, ".env")
	}

	if t == "config" {
		return filepath.Join(datadir, "nigiri.config.json")
	}

	return ""
}

func nigiriIsRunning() (bool, error) {
	listOnlyRunningContainers := !listAll
	return nigiriExists(listOnlyRunningContainers)
}

func nigiriExistsAndNotRunning() (bool, error) {
	return nigiriExists(listAll)
}

func getStartBashCmd(datadir string) (*exec.Cmd, error) {
	composePath := getPath(datadir, "compose")
	envPath := getPath(datadir, "env")
	env := loadEnv(envPath)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")

	isStopped, err := nigiriExistsAndNotRunning()
	if err != nil {
		return nil, err
	}
	if isStopped {
		bashCmd = exec.Command("docker-compose", "-f", composePath, "start")
	}
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Env = env

	return bashCmd, nil
}

func writeComposeEnvFile(path string, stringifiedJSON string) error {
	defaultJSON, _ := json.Marshal(defaultPorts)
	env := map[string]map[string]int{}
	json.Unmarshal([]byte(stringifiedJSON), &env)

	if stringifiedJSON != string(defaultJSON) {
		env = mergeComposeEnvFiles([]byte(stringifiedJSON))
	}

	fileContent := ""
	for chain, services := range env {
		for k, v := range services {
			fileContent += fmt.Sprintf("%s_%s_PORT=%d\n", strings.ToUpper(chain), strings.ToUpper(k), v)
		}
	}

	return ioutil.WriteFile(path, []byte(fileContent), os.ModePerm)
}

func readComposeEnvFile(path string) (map[string]map[string]int, error) {
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
	// Each line is in the format PREFIX_SERVICE_NAME_SUFFIX=value
	// PREFIX is either 'BITCOIN' or 'LIQUID', while SUFFIX is always 'PORT'
	for scanner.Scan() {
		line := scanner.Text()
		splitLine := strings.Split(line, "=")
		key := splitLine[0]
		value, _ := strconv.Atoi(splitLine[1])
		chain := "bitcoin"
		if strings.HasPrefix(key, strings.ToUpper("liquid")) {
			chain = "liquid"
		}

		suffix := "_PORT"
		prefix := strings.ToUpper(fmt.Sprintf("%s_", chain))
		trimmedKey := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(key, prefix), suffix))
		ports[chain][trimmedKey] = value
	}

	return ports, nil
}

func mergeComposeEnvFiles(rawJSON []byte) map[string]map[string]int {
	newPorts := map[string]map[string]int{}
	json.Unmarshal(rawJSON, &newPorts)

	mergedPorts := map[string]map[string]int{}
	for chain, services := range defaultPorts {
		mergedPorts[chain] = make(map[string]int)
		for name, port := range services {
			newPort := newPorts[chain][name]
			if newPort > 0 && newPort != port {
				mergedPorts[chain][name] = newPort
			} else {
				mergedPorts[chain][name] = port
			}
		}
	}

	return mergedPorts
}

func loadEnv(path string) []string {
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
