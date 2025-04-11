package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	lib "github.com/charles-m-knox/finance-planner-lib"
	"github.com/charles-m-knox/go-uuid"

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

// fileExists simply checks if a file exists. Nothing too fancy.
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

// Attempts to load from the "file" path provided. If unsuccessful, an example
// file will be loaded, and the second argument will be true.
//
// The "t" parameter is the map of translations.
func loadConfig(file string, t map[string]string, exampleConf embed.FS) (Config, bool, error) {
	if file == "" {
		fmt.Println(t["ConfigNoFileSpecified"])
		os.Exit(1)
	}

	var err error

	var exists bool

	var conf Config

	exists, err = fileExists(file)
	if err != nil {
		return conf, false, fmt.Errorf("failed to check if file %v exists: %w ", file, err)
	}

	if exists {
		conf, file, err = loadConfFrom(file, t)
		if err != nil {
			return conf, false, fmt.Errorf("failed to load config from existing config file %v: %w ", file, err)
		}

		return conf, false, nil
	}

	// if it doesn't exist, use an example
	conf, file, err = loadConfFromEmbed("example.yml", exampleConf, t)
	if err != nil {
		return conf, false, fmt.Errorf("failed to load config from template config %v: %w ", file, err)
	}

	return conf, true, err
}

// processConfig applies any post-load configuration parameters/logic to ensure
// that data is valid & consistent. Use it after loadConfig.
func processConfig(conf *Config) {
	if conf == nil {
		log.Fatalf("config is nil")
	}

	// ensure that every transaction has its weekdays map properly populated
	for i := range 7 {
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
