package main

import (
	"flag"
	"fmt"
	skos2web "github.com/antleaf/skos2web"
	"go.uber.org/zap"
	"path/filepath"
)

var zapLogger *zap.Logger
var config = Config{}
var conceptSchemes []*skos2web.ConceptScheme
var website = skos2web.Website{}

func main() {
	var err error
	var configFilePath string
	zapLogger, _ = configureZapLogger(false)
	//	### Load arguments
	flag.StringVar(&configFilePath, "c", "", "path to a valid config yaml file")
	flag.Parse()

	// ### Load configuration
	err = (&config).unmarshal(configFilePath)
	if err != nil {
		zapLogger.Error(err.Error())
		zapLogger.Fatal("Unable to initialise - halting execution")
	} else {
		zapLogger.Info("Configuration loaded OK")
	}
	// ### Load SKOS configs and initialise conceptSchemes
	conceptSchemes = make([]*skos2web.ConceptScheme, 0)
	for _, conceptSchemeConfig := range config.ConceptSchemeConfigs {
		conceptSchemePtr := new(skos2web.ConceptScheme)
		err = conceptSchemePtr.Initialise(conceptSchemeConfig.ID, conceptSchemeConfig.Title, conceptSchemeConfig.Uri, conceptSchemeConfig.Namespace, conceptSchemeConfig.Description, conceptSchemeConfig.Version, conceptSchemeConfig.SkosSourceFolderPath, conceptSchemeConfig.Creators, conceptSchemeConfig.Contributors, conceptSchemeConfig.Updated)
		if err != nil {
			zapLogger.Error(err.Error())
			zapLogger.Fatal(fmt.Sprintf("Unable to initialise concept scheme '%s' - halting immediately", conceptSchemeConfig.ID))
		}
		dSpaceXmlFilePath := filepath.Join(conceptSchemePtr.SkosProcessedFolderPath, conceptSchemePtr.ID+"_for_dspace.xml")
		err = generateDspaceXml(conceptSchemePtr, dSpaceXmlFilePath)
		if err != nil {
			zapLogger.Error(err.Error())
			zapLogger.Fatal(fmt.Sprintf("Unable to create DSpace XML for concept scheme '%s' - halting immediately", conceptSchemeConfig.ID))
		}
		conceptSchemes = append(conceptSchemes, conceptSchemePtr)
	}
	zapLogger.Info(fmt.Sprintf("%v Concept schemes created and processed OK", len(config.ConceptSchemeConfigs)))
	// ### Initialise Hugo content and static folders
	(&website).Initialise(config.WebrootFolderPath)
	zapLogger.Info("Building website...")
	for _, conceptScheme := range conceptSchemes {
		err = website.ProcessConceptScheme(conceptScheme)
		if err != nil {
			zapLogger.Error(err.Error())
			zapLogger.Fatal("Unable to process vocabulary for website - halting immediately")
		}
		err = website.GenerateZip(conceptScheme, []string{conceptScheme.WorkingFilePathNTriples, filepath.Join(conceptScheme.SkosProcessedFolderPath, conceptScheme.ID+"_for_dspace.xml")})
		if err != nil {
			zapLogger.Error(err.Error())
			zapLogger.Fatal("Unable to generate zip file for website - halting immediately")
		}
	}
	zapLogger.Info("Website built OK")
	zapLogger.Info("Process completed successfully")
}
