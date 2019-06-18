package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/altafan/nigiri/cli/config"

	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "Check service logs",
	RunE:    logs,
	PreRunE: logsChecks,
}

var services = map[string]bool{
	"node":       true,
	"electrs":    true,
	"esplora":    true,
	"chopsticks": true,
}

func logsChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	if !isDatadirOk(datadir) {
		return fmt.Errorf("Invalid datadir, it must be an absolute path: %s", datadir)
	}
	if len(args) != 1 {
		return fmt.Errorf("Invalid number of args, expected 1, got: %d", len(args))
	}

	service := args[0]
	if !services[service] {
		return fmt.Errorf("Invalid service, must be one of %s. Got: %s", reflect.ValueOf(services).MapKeys(), service)
	}
	isRunning, err := nigiriIsRunning()
	if err != nil {
		return err
	}
	if !isRunning {
		return fmt.Errorf("Nigiri is not running")
	}

	if err := config.ReadFromFile(datadir); err != nil {
		return err
	}

	if isLiquidService && isLiquidService != config.GetBool(config.AttachLiquid) {
		return fmt.Errorf("Nigiri has been started with no Liquid sidechain.\nPlease stop and restart it using the --liquid flag")
	}

	return nil
}

func logs(cmd *cobra.Command, args []string) error {
	service := args[0]
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	serviceName := getServiceName(service, isLiquidService)
	composePath := getPath(datadir, "compose")
	envPath := getPath(datadir, "env")
	env := loadEnv(envPath)

	bashCmd := exec.Command("docker-compose", "-f", composePath, "logs", serviceName)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Env = env

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}

func getServiceName(name string, liquid bool) string {
	service := name
	if service == "node" {
		service = "bitcoin"
	}
	if liquid {
		if service == "bitcoin" {
			service = "liquid"
		} else {
			service = fmt.Sprintf("%s-liquid", service)
		}
	}

	return service
}
