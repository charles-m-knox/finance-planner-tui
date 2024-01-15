package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	c "gitea.cmcode.dev/cmcode/finance-planner-tui/constants"
	"gitea.cmcode.dev/cmcode/finance-planner-tui/lib"
	m "gitea.cmcode.dev/cmcode/finance-planner-tui/models"

	"github.com/adrg/xdg"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Attempts to load from a specific location, if possible.
//
// The first return value is the populated config, if one was found and parsed.
// The second return value is a string that indicates the properly loaded path
// that successfully loaded the config (if it didn't succeed, it will be an
// empty string). The third return value is an error, if present.
//
// The "t" parameter is the map of translations.
func loadConfFrom(file string, t map[string]string) (m.Config, string, error) {
	conf := m.Config{}

	b, err := os.ReadFile(file)
	if err != nil {
		return conf, "", fmt.Errorf("%v %v: %w", t["ConfigFailedToLoadConfig"], file, err)
	}

	err = yaml.Unmarshal(b, &conf)

	if err != nil {
		return conf, "", fmt.Errorf("%v %v: %w", t["ConfigFailedToUnmarshalConfig"], file, err)
	}

	return conf, file, nil
}

// Attempts to load from the "file" path provided - if not successful,
// attempts to load from xdg config, then xdg home.
//
// The first return value is the populated config, if one was found and parsed.
// The second return value is a string that indicates the properly loaded path
// that successfully loaded the config (if it didn't succeed, it will be an
// empty string). The third return value is an error, if present.
//
// You should set the global configFile variable to match the returned string
// value so that other logic can use it.
//
// The "t" parameter is the map of translations.
func loadConfig(file string, t map[string]string) (m.Config, string, error) {
	if file == "" {
		file = c.DefaultConfig
	}

	var err error

	var conf m.Config

	conf, file, err = loadConfFrom(file, t)

	if err == nil && file != "" {
		return conf, file, err
	} else if err != nil {
		return conf, file, err
	}

	xdgConfig := path.Join(xdg.ConfigHome, c.DefaultConfigParentDir, c.DefaultConfig)

	conf, file, err = loadConfFrom(xdgConfig, t)
	if err == nil && file != "" {
		return conf, file, err
	}

	xdgHome := path.Join(xdg.Home, c.DefaultConfigParentDir, c.DefaultConfig)

	conf, file, err = loadConfFrom(xdgHome, t)
	if err == nil && file != "" {
		return conf, file, err
	} else if err != nil {
		return conf, file, err
	}

	// if the config file doesn't exist, create it at xdgConfig
	return conf, xdgConfig, err
}

// processConfig applies any post-load configuration parameters/logic to ensure
// that data is valid & consistent. Use it after loadConfig.
func processConfig(conf *m.Config) {
	if conf == nil {
		log.Fatalf("config is nil")
	}

	// ensure that every transaction has its weekdays map properly populated
	for i := 0; i < 7; i++ {
		for j := range conf.Profiles {
			for k := range conf.Profiles[j].TX {
				_, ok := conf.Profiles[j].TX[k].Weekdays[i]
				if !ok {
					conf.Profiles[j].TX[k].Weekdays[i] = false
				}
			}
		}
	}
}

// converts a json file to yaml (one-off job for converting from legacy versions
// of this program).
func JSONtoYAML() {
	b, err := os.ReadFile("conf.json")
	if err != nil {
		log.Fatalf("failed to load conf.json")
	}

	nc := m.Config{
		Profiles: []m.Profile{
			{
				Name: "migrated",
				TX:   []lib.TX{},
			},
		},
	}

	err = json.Unmarshal(b, &nc.Profiles[0].TX)
	if err != nil {
		log.Fatalf("failed to unmarshal conf: %v", err.Error())
	}

	// update all uuids in the config
	for i := range nc.Profiles[0].TX {
		nc.Profiles[0].TX[i].ID = uuid.NewString()
	}

	out, err := yaml.Marshal(nc)
	if err != nil {
		log.Fatalf("failed to marshal nc: %v", err.Error())
	}

	err = os.WriteFile("migrated.yml", out, 0o644)
	if err != nil {
		log.Fatalf("failed to write migrated.yml: %v", err.Error())
	}
}
