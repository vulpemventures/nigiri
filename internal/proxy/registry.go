package proxy

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/sdomino/scribble"
)

// Registry handles writing/reading json file where are stored info about issued assets
type Registry struct {
	db *scribble.Driver
}

// NewRegistry returns a new Registry or error if the path is not absolute
func NewRegistry(path string) (*Registry, error) {
	r := &Registry{}
	db, err := scribble.New(path, nil)
	if err != nil {
		return nil, err
	}
	r.db = db
	return r, nil
}

// AddEntry adds an entry to the register by previously making sure that
// the incoming entry does not already exist in the registry
func (r *Registry) AddEntry(asset string, issuanceInput map[string]interface{}, contract map[string]interface{}) error {
	entry := map[string]interface{}{}
	err := r.db.Read("registry", asset, &entry)
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return err
	}

	if len(entry) != 0 {
		return errors.New("Asset already exists on registry")
	}

	entry = map[string]interface{}{
		"asset":         asset,
		"issuance_txin": issuanceInput,
		"contract":      contract,
		"name":          contract["name"].(string),
		"ticker":        contract["ticker"].(string),
		"precision":     contract["precision"].(int),
	}
	return r.db.Write("registry", asset, entry)
}

// GetEntry returns an entry if it exists in registry or NIL
func (r *Registry) GetEntry(asset string) (map[string]interface{}, error) {
	entry := map[string]interface{}{}
	err := r.db.Read("registry", asset, &entry)
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return nil, err
	}

	return entry, nil
}

// GetEntries returns multiple entries from the registry
func (r *Registry) GetEntries(assets []interface{}) ([]map[string]interface{}, error) {
	entries := []map[string]interface{}{}

	if len(assets) == 0 {
		records, err := r.db.ReadAll("registry")
		if err != nil {
			return nil, err
		}

		for _, f := range records {
			entry := map[string]interface{}{}
			json.Unmarshal([]byte(f), &entry)
			entries = append(entries, entry)
		}
	} else {
		for _, asset := range assets {
			entry := map[string]interface{}{}
			r.db.Read("registry", asset.(string), &entry)
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
