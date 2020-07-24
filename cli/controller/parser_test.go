package controller

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/vulpemventures/nigiri/cli/constants"
)

func TestParserParseNetwork(t *testing.T) {
	p := &Parser{}

	validNetworks := []string{"regtest"}
	for _, n := range validNetworks {
		err := p.parseNetwork(n)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParserParseDatadir(t *testing.T) {
	p := &Parser{}

	currentDir, _ := os.Getwd()
	validDatadirs := []string{currentDir}
	for _, n := range validDatadirs {
		err := p.parseDatadir(n)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParserParseEnvJSON(t *testing.T) {
	p := &Parser{}

	for _, e := range testJSONs {
		parsedJSON, _ := json.Marshal(e)
		mergedJSON, err := p.parseEnvJSON(string(parsedJSON))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(mergedJSON)
	}
}

func TestParserParseNetworkShouldFail(t *testing.T) {
	p := &Parser{}

	invalidNetworks := []string{"simnet", "testnet"}
	for _, n := range invalidNetworks {
		err := p.parseNetwork(n)
		if err == nil {
			t.Fatalf("Should have been failed before")
		}
		if err != constants.ErrInvalidNetwork {
			t.Fatalf("Got: %s, wanted: %s", err, constants.ErrInvalidNetwork)
		}
	}
}

func TestParserParseDatadirShouldFail(t *testing.T) {
	p := &Parser{}

	invalidDatadirs := []string{"."}
	for _, d := range invalidDatadirs {
		err := p.parseDatadir(d)
		if err == nil {
			t.Fatalf("Should have been failed before")
		}
		if err != constants.ErrInvalidDatadir {
			t.Fatalf("Got: %s, wanted: %s", err, constants.ErrInvalidDatadir)
		}
	}
}

var testJSONs = []map[string]interface{}{
	// only btc services
	{
		"urls": map[string]string{
			"bitcoin_esplora": "https://blockstream.info/",
		},
		"ports": map[string]map[string]int{
			"bitcoin": map[string]int{
			    "peer":       1111,
				"node":       2222,
				"esplora":    3333,
				"electrs":    4444,
				"chopsticks": 5555,
			},
		},
	},
	// btc and liquid services
	{
		"urls": map[string]string{
			"bitcoin_esplora": "https://blockstream.info/",
			"liquid_esplora":  "http://blockstream.info/liquid",
		},
		"ports": map[string]map[string]int{
			"bitcoin": map[string]int{
				"peer":       1111,
                "node":       2222,
                "esplora":    3333,
                "electrs":    4444,
                "chopsticks": 5555,
			},
			"liquid": map[string]int{
				"peer":       6666,
				"node":       7777,
				"esplora":    8888,
				"electrs":    9999,
				"chopsticks": 1010,
			},
		},
	},
	// incomplete examples:
	// incomplete bitcoin services
	{
		"ports": map[string]map[string]int{
			"bitcoin": map[string]int{
				"esplora":    1111,
				"electrs":    2222,
				"chopsticks": 3333,
			},
		},
		"urls": map[string]string{
			"bitcoin_esplora": "http://test.com/api",
		},
	},
	// bitcoin services ports and liquid service url
	{
		"ports": map[string]map[string]int{
			"bitcoin": map[string]int{
				"node":       1111,
				"esplora":    2222,
				"electrs":    3333,
				"chopsticks": 4444,
			},
		},
		"urls": map[string]string{
			"liquid_esplora": "http://test.com/liquid/api",
		},
	},
	// liquid services ports and bitcoin service url
	{
		"ports": map[string]map[string]int{
			"liquid": map[string]int{
				"node":       1111,
				"esplora":    2222,
				"electrs":    3333,
				"chopsticks": 4444,
			},
		},
		"urls": map[string]string{
			"bitcoin_esplora": "http://test.com/api",
		},
	},
	// empty config
	{},
}
