package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"time"
)

type ConceptSchemeVersionConfig struct {
	VersionNumber        string                     `yaml:"version"`
	SkosSourceFolderPath string                     `yaml:"folder_path"`
	Details              *ConceptSchemeDetailConfig `yaml:"-"`
}

type ConceptSchemeConfig struct {
	ID       string                        `yaml:"name"`
	Versions []*ConceptSchemeVersionConfig `yaml:"versions"`
}

type ConceptSchemeDetailConfig struct {
	Title        string    `yaml:"title"`
	Description  string    `yaml:"description"`
	Namespace    string    `yaml:"namespace"`
	Updated      time.Time `yaml:"updated"`
	Creators     []string  `yaml:"creators"`
	Contributors []string  `yaml:"contributors"`
}

type Config struct {
	Debugging                   bool                   `yaml:"debugging"`
	WebrootFolderPath           string                 `yaml:"webroot"`
	ProcessedSkosRootFolderPath string                 `yaml:"processed_skos_root"`
	ConceptSchemeConfigs        []*ConceptSchemeConfig `yaml:"concept_schemes"`
}

func (config *Config) unmarshal(filePath string) error {
	configData, err := ioutil.ReadFile(filePath)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	//config.ConceptSchemeConfigs = make([]ConceptSchemeConfig, 0)
	err = yaml.Unmarshal([]byte(configData), config)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	if config.Debugging == true {
		zapLogger, _ = configureZapLogger(true)
		EnableDebugging()
		zapLogger.Info("Debugging enabled")
	}
	for _, conceptSchemeConfig := range config.ConceptSchemeConfigs {
		for _, conceptSchemeVersionConfig := range conceptSchemeConfig.Versions {
			conceptSchemeVersionDetails, err := ioutil.ReadFile(filepath.Join(conceptSchemeVersionConfig.SkosSourceFolderPath, "config.yaml"))
			if err != nil {
				zapLogger.Error(err.Error())
				return err
			}
			var conceptSchemeVersionDetailsConfig = ConceptSchemeDetailConfig{}
			err = yaml.Unmarshal([]byte(conceptSchemeVersionDetails), &conceptSchemeVersionDetailsConfig)
			if err != nil {
				zapLogger.Error(err.Error())
				return err
			}
			conceptSchemeVersionConfig.Details = &conceptSchemeVersionDetailsConfig
		}

	}
	zapLogger.Info(fmt.Sprintf("%v Concept Scheme configurations loaded", len(config.ConceptSchemeConfigs)))
	zapLogger.Info(fmt.Sprintf("Webroot folder path set to %s", config.WebrootFolderPath))
	return err
}