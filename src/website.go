package main

import (
	"fmt"
	"github.com/goki/ki/ki"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Website struct {
	WebrootFolderPath       string
	ContentFolderPath       string
	StaticContentFolderPath string
}

func (website *Website) Initialise(webrootPath string) error {
	var err error
	website.WebrootFolderPath = webrootPath
	website.ContentFolderPath = filepath.Join(website.WebrootFolderPath, "content")
	website.StaticContentFolderPath = filepath.Join(website.WebrootFolderPath, "static")
	err = os.MkdirAll(website.ContentFolderPath, os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = os.MkdirAll(website.StaticContentFolderPath, os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}

func (website *Website) ProcessConceptSchemeVersion(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	err = os.MkdirAll(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	conceptSchemePage, conceptSchemeMarshalErr := conceptSchemeVersion.Marshal()
	if conceptSchemeMarshalErr != nil {
		zapLogger.Error(conceptSchemeMarshalErr.Error())
		return conceptSchemeMarshalErr
	}
	fileWriteErr := ioutil.WriteFile(filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), "_index.md"), conceptSchemePage, os.ModePerm)
	if fileWriteErr != nil {
		zapLogger.Error(fileWriteErr.Error())
		return fileWriteErr
	}
	err = website.GenerateConceptPages(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = website.GenerateHtmlTree(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = website.GeneratePrintableSinglePage(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = website.GenerateZip(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}

func (website *Website) GeneratePrintableSinglePage(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	conceptSchemeVersion.HugoLayout = "printable"
	printableVocabPage, conceptSchemeMarshalErr := conceptSchemeVersion.Marshal()
	if conceptSchemeMarshalErr != nil {
		zapLogger.Error(conceptSchemeMarshalErr.Error())
		return conceptSchemeMarshalErr
	}
	conceptSchemeVersion.HugoLayout = ""
	err = ioutil.WriteFile(filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), "printable.md"), printableVocabPage, os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}

func (website *Website) GenerateConceptPages(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	for _, concept := range conceptSchemeVersion.Concepts {
		conceptPage, conceptMarshalErr := concept.marshal()
		if conceptMarshalErr != nil {
			zapLogger.Error(conceptMarshalErr.Error())
			return conceptMarshalErr
		}
		conceptPageFolderPath := filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), concept.ID)
		os.MkdirAll(conceptPageFolderPath, os.ModePerm)
		conceptFileWriteErr := ioutil.WriteFile(filepath.Join(conceptPageFolderPath, "index.md"), conceptPage, os.ModePerm)
		if conceptFileWriteErr != nil {
			zapLogger.Error(conceptFileWriteErr.Error())
			return conceptFileWriteErr
		}
	}
	return err
}

func (website *Website) GenerateZip(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	filesPathsToZip := []string{conceptSchemeVersion.WorkingFilePathNTriples, filepath.Join(conceptSchemeVersion.SkosProcessedFolderPath, conceptSchemeVersion.ID+"_for_dspace.xml")}
	err := zipFiles(filepath.Join(website.StaticContentFolderPath, fmt.Sprint(conceptSchemeVersion.ID, "_", conceptSchemeVersion.VersionNumberString, ".zip")), filesPathsToZip)
	if err != nil {
		zapLogger.Debug(err.Error())
		return err
	}
	return err
}

func (website *Website) GenerateHtmlTree(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	html := "<ul id=\"tree-root\">"
	treeDepth := 0
	finalNodeDepth := 0
	var treeFunction ki.Func
	treeFunction = func(k ki.Ki, level int, data interface{}) bool {
		concept := conceptSchemeVersion.GetConceptById(k.Name())
		finalNodeDepth = level
		if level > treeDepth {
			html += "<ul>"
		} else if level < treeDepth {
			stepsBack := treeDepth - level
			for i := 1; i <= stepsBack; i++ {
				html += "</ul>"
			}
		}
		treeDepth = level
		if k.Name() != conceptSchemeVersion.ID {
			if asCurrentVersion {
				html += fmt.Sprintf("<li><a href=\"/%s/%s/\">%s</a></li>", conceptSchemeVersion.ID, concept.ID, concept.Title)
			} else {
				html += fmt.Sprintf("<li><a href=\"/%s/%s/%s/\">%s</a></li>", conceptSchemeVersion.ID, conceptSchemeVersion.VersionNumberString, concept.ID, concept.Title)
			}
		}
		return true
	}
	conceptSchemeVersion.FuncDownMeFirst(treeDepth, conceptSchemeVersion, treeFunction)
	for i := 0; i <= finalNodeDepth; i++ {
		html += "</ul>"
	}
	err = ioutil.WriteFile(filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), "tree.txt"), []byte(html), os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}
