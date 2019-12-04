package controller

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/vulpemventures/nigiri/cli/constants"
)

// Parser implements functions for parsing flags, JSON files
// and system directories passed to the CLI commands
type Parser struct {
	services map[string]bool
}

func newParser(services map[string]bool) *Parser {
	p := &Parser{services}
	return p
}

func (p *Parser) parseNetwork(network string) error {
	for _, n := range constants.AvaliableNetworks {
		if network == n {
			return nil
		}
	}
	return constants.ErrInvalidNetwork
}

func (p *Parser) parseDatadir(path string) error {
	if !filepath.IsAbs(path) {
		return constants.ErrInvalidDatadir
	}
	return nil
}

func (p *Parser) parseEnvJSON(strJSON string) (string, error) {
	// merge default json and incoming json by parsing DefaultEnv to
	// envJSON type and then parsing the incoming json using the same variable
	var parsedJSON envJSON
	defaultJSON, _ := json.Marshal(constants.DefaultEnv)
	json.Unmarshal(defaultJSON, &parsedJSON)
	err := json.Unmarshal([]byte(strJSON), &parsedJSON)
	if err != nil {
		fmt.Println(err)
		return "", constants.ErrMalformedJSON
	}
	merged, _ := json.Marshal(parsedJSON)
	return string(merged), nil
}

func (p *Parser) parseServiceName(name string) error {
	if !p.services[name] {
		return constants.ErrInvalidServiceName
	}
	return nil
}
