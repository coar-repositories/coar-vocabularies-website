package main

import (
	"fmt"
	skos2web "github.com/antleaf/skos2web"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/goki/ki/ki"
	"io/ioutil"
	"os"
	"strings"
)

func generateDspaceXml(conceptScheme *skos2web.ConceptScheme, dSpaceXmlFilePath string) error {
	zapLogger.Debug(fmt.Sprintf("Creating DSpace XML file for '%s' at %s", conceptScheme.ID, dSpaceXmlFilePath))
	var err error
	xml := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	xml += "<!-- This XML was automatically generated from the SKOS sources for the COAR Vocabulary. It follows the schema developed by 4Science (https://www.4science.it/) for DSpace -->"
	currentLevel := 0
	finalNodeDepth := 0
	conceptScheme.FuncDownMeFirst(0, nil, func(k ki.Ki, level int, d interface{}) bool {
		xmlNode := ""
		if k.Name() == conceptScheme.ID {
			xmlNode = fmt.Sprintf("<node id=\"%s\" label=\"%s\">", conceptScheme.ID, conceptScheme.Title)
		} else {
			concept := conceptScheme.GetConceptById(k.Name())
			xmlNode = fmt.Sprintf("<node id=\"%s\" label=\"%s\">", concept.ID, concept.Title)
			if concept.Definition != "" {
				xmlNode += fmt.Sprintf("<hasNote>%s</hasNote>", concept.Definition)
			}
		}
		if level < currentLevel {
			for i := 0; i < (currentLevel - level); i++ {
				xml += "</isComposedBy></node>"
			}
		}
		xml += xmlNode
		if k.HasChildren() {
			xml += "<isComposedBy>"
		} else {
			xml += "</node>"
		}
		currentLevel = level
		finalNodeDepth = level
		return true // return value determines whether tree traversal continues or not
	})
	for i := 0; i < finalNodeDepth; i++ {
		xml += "</isComposedBy></node>"
	}
	xml = xmlfmt.FormatXML(xml, "", "  ")
	xml = strings.TrimSpace(xml)
	err = ioutil.WriteFile(dSpaceXmlFilePath, []byte(xml), os.ModePerm)
	return err
}
