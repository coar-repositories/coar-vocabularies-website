package main

import (
	"fmt"
	"github.com/antleaf/skos2web"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type ConceptSchemeConfig struct {
	ID                   string    `yaml:"name"`
	SkosSourceFolderPath string    `yaml:"skos_folder_path"`
	Version              string    `yaml:"version"`
	Title                string    `yaml:"title"`
	Description          string    `yaml:"description"`
	Uri                  string    `yaml:"uri"`
	Namespace            string    `yaml:"namespace"`
	Updated              time.Time `yaml:"updated"`
	Creators             []string  `yaml:"creators"`
	Contributors         []string  `yaml:"contributors"`
}

type Config struct {
	Debugging            bool                  `yaml:"debugging"`
	WebrootFolderPath    string                `yaml:"webroot"`
	ConceptSchemeConfigs []ConceptSchemeConfig `yaml:"concept_schemes"`
}

func (config *Config) unmarshal(filePath string) error {
	configData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	config.ConceptSchemeConfigs = make([]ConceptSchemeConfig, 0)
	err = yaml.Unmarshal([]byte(configData), config)
	if err != nil {
		return err
	}
	if config.Debugging == true {
		zapLogger, _ = configureZapLogger(true)
		skos2web.EnableDebugging()
		zapLogger.Info("Debugging enabled")
	}
	zapLogger.Info(fmt.Sprintf("Webroot folder path set to %s", config.WebrootFolderPath))
	return err
}
