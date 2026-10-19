package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

// devin's catalog is one JSON document the command prints and exits on, not
// a server to talk to. Its families group the variants --model takes, so
// each variant's model_uid is the id a session launches with.
func readDevin(ctx context.Context, command, dir string) (Catalog, error) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return Catalog{}, errors.New("empty catalog command")
	}
	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return Catalog{}, err
	}
	var listed struct {
		Families []struct {
			Variants []struct {
				ID    string `json:"model_uid"`
				Label string `json:"label"`
			} `json:"variants"`
		} `json:"families"`
	}
	if err := json.Unmarshal(out, &listed); err != nil {
		return Catalog{}, err
	}
	var cat Catalog
	for _, family := range listed.Families {
		for _, variant := range family.Variants {
			cat.Models = append(cat.Models, Model{ID: variant.ID, Label: variant.Label})
		}
	}
	return cat, nil
}
