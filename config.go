package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	lib "github.com/charles-m-knox/finance-planner-lib"
	"github.com/charles-m-knox/go-uuid"

	"github.com/adrg/xdg"
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
func loadConfFrom(file string, t map[string]string) (Config, string, error) {
	conf := Config{}

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

func loadConfFromEmbed(file string, emb embed.FS, t map[string]string) (Config, string, error) {
	conf := Config{}

	b, err := emb.ReadFile(file)
	if err != nil {
		return conf, "", fmt.Errorf("%v %v: %w", t["ConfigFailedToLoadEmbeddedConfig"], file, err)
	}

	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return conf, "", fmt.Errorf("%v %v: %w", t["ConfigFailedToUnmarshalEmbeddedConfig"], file, err)
	}

	return conf, file, nil
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
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
func loadConfig(file string, t map[string]string, exampleConf embed.FS) (Config, string, error) {
	if file == "" {
		file = DefaultConfig
	}

	var err error

	var exists bool

	var conf Config

	// create the XDG config dir for this application once upon startup
	xdgConfigDir := path.Join(xdg.ConfigHome, DefaultConfigParentDir)

	err = os.MkdirAll(xdgConfigDir, 0o755)
	if err != nil {
		return conf, file, fmt.Errorf("failed to make all directories %v: %w ", xdgConfigDir, err)
	}

	exists, err = fileExists(file)
	if err != nil {
		return conf, file, fmt.Errorf("failed to check if file %v exists: %w ", file, err)
	}

	if exists {
		conf, file, err = loadConfFrom(file, t)
		if err != nil {
			return conf, file, fmt.Errorf("failed to load config from existing config file %v: %w ", file, err)
		}

		return conf, file, nil
	}

	xdgConfig := path.Join(xdgConfigDir, DefaultConfig)

	exists, err = fileExists(xdgConfig)
	if err != nil {
		return conf, file, fmt.Errorf("failed to check if file %v exists: %w ", file, err)
	}

	if exists {
		conf, file, err = loadConfFrom(xdgConfig, t)
		if err != nil {
			return conf, file, fmt.Errorf("failed to load config from existing config file %v: %w ", file, err)
		}

		return conf, file, nil
	}

	xdgHome := path.Join(xdg.Home, DefaultConfigParentDir, DefaultConfig)

	exists, err = fileExists(xdgHome)
	if err != nil {
		return conf, file, fmt.Errorf("failed to check if file %v exists: %w ", file, err)
	}

	if exists {
		conf, file, err = loadConfFrom(xdgConfig, t)
		if err != nil {
			return conf, file, fmt.Errorf("failed to load config from existing config file %v: %w ", file, err)
		}

		return conf, file, nil
	}

	// if the config file doesn't exist, create it at xdgConfig with the
	// example config (note: this doesn't *write* to the xdgConfig path,
	// but instead sets the target config write path there so that it will
	// be saved there)
	conf, file, err = loadConfFromEmbed("example.yml", exampleConf, t)
	if err != nil {
		return conf, file, fmt.Errorf("failed to load config from template config %v: %w ", file, err)
	}

	return conf, xdgConfig, err
}

// processConfig applies any post-load configuration parameters/logic to ensure
// that data is valid & consistent. Use it after loadConfig.
func processConfig(conf *Config) {
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

	nc := Config{
		Profiles: []Profile{
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
		nc.Profiles[0].TX[i].ID = uuid.New()
	}

	out, err := yaml.Marshal(nc)
	if err != nil {
		log.Fatalf("failed to marshal nc: %v", err.Error())
	}

	//nolint:gosec
	err = os.WriteFile("migrated.yml", out, 0o644)
	if err != nil {
		log.Fatalf("failed to write migrated.yml: %v", err.Error())
	}
}
