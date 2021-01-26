/*
Package config provides tools for managing the configuration file.
*/
package config

import (
	"errors"
	"io/ioutil"
	"os"

	"Prefab-Catalog/lib/lumberjack"

	"gopkg.in/yaml.v3"
)

/*
VARIABLES
*/

var log = lumberjack.New("Config")

var defaultConfigString string = `DatabaseURI: URI
DatabaseTimeout: 10
LogPath: /root/log/
Port: 80
Verbosity: 1
URL: http://example.com`

var defaultConfig []byte = []byte(defaultConfigString)

func init() {
	_, err := os.Stat("config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			err := ioutil.WriteFile("config.yaml", defaultConfig, 0755)
			if err != nil {
				log.Fatal(err, "Unable to create configuration file.")
			} else {
				err = errors.New("")
				log.Fatal(err, "A new configuration file was created. Enter required values.")
			}
		} else {
			log.Fatal(err, "Unable to read configuration file")
		}
	}
}

// YAML describes the structure of the YAML configuration file. See https://godoc.org/gopkg.in/yaml.v3#Marshal for more information.
type YAML struct {
	DatabaseURI     string `yaml:"DatabaseURI"`
	DatabaseTimeout int    `yaml:"DatabaseTimeout"`
	UploadPath      string `yaml:"UploadPath"`
	LogPath         string `yaml:"LogPath"`
	Port            int    `yaml:"Port"`
	Verbosity       int    `yaml:"Verbosity"`
	URL             string `yaml:"URL"` // The URL the prod server is running at.
}

/*
Build returns the build as a string.
*/
func Build() string {
	return "0.2 (2021-01-26)  " // TODO: Update build number each time a push is made to Git).
}

/*
Load configuration values from "config.yaml".
*/
func Load() YAML {
	configuration, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err, "Unable to read configuration file")
	}

	var config YAML
	err2 := yaml.Unmarshal(configuration, &config)
	if err2 != nil {
		log.Fatal(err2, "Unable to parse configuration file as YAML")
	}

	// Parse configuration settings.

	// Check mandatory settings.
	err = errors.New("required setting missing in configuration file")
	if config.DatabaseURI == "" {
		log.Fatal(err, "DatabaseURI")
	}
	if config.URL == "" {
		log.Warn("No server URL specified in config -- mail sent will not have a link to the order record.")
		config.URL = "localhost/"
	}
	if config.DatabaseTimeout == 0 {
		config.DatabaseTimeout = 10
	}
	if config.LogPath == "" {
		config.LogPath = "./logs/"
	}
	if config.UploadPath == "" {
		config.UploadPath = "./upload/"
	}
	return config
}
