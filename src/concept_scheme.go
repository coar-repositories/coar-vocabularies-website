package main

import (
	"errors"
	"fmt"
	"github.com/knakk/rdf"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type ConceptScheme struct {
	ID       string
	Versions []ConceptSchemeVersion
}

func (conceptScheme *ConceptScheme) Initialise(config *ConceptSchemeConfig, processedSkosRootFolderPath string) error {
	zapLogger.Debug(fmt.Sprintf("Initialising concept scheme: '%s'", config.ID))
	var err error
	conceptScheme.ID = config.ID
	for _, versionConfig := range config.Versions {
		zapLogger.Debug("Title", zap.String("title", versionConfig.Details.Title))
		var version = ConceptSchemeVersion{
			ID:                  conceptScheme.ID,
			VersionNumberString: versionConfig.VersionNumber,
			Title:               versionConfig.Details.Title,
			Description:         versionConfig.Details.Description,
			Namespace:           versionConfig.Details.Namespace,
			Updated:             versionConfig.Details.Updated,
			Creators:            versionConfig.Details.Creators,
			Contributors:        versionConfig.Details.Contributors,
			HugoLayout:          "",
		}
		version.VersionNumber, err = strconv.ParseFloat(version.VersionNumberString, 64)
		if err != nil {
			zapLogger.Error("Unable to convert version number to string" + err.Error())
			return err
		}
		version.SkosProcessedFolderPath = filepath.Join(processedSkosRootFolderPath, conceptScheme.ID, version.VersionNumberString)
		err = resetVolatileFolder(version.SkosProcessedFolderPath)
		if err != nil {
			zapLogger.Error(err.Error())
			return err
		}
		version.WorkingFilePathNTriples = filepath.Join(version.SkosProcessedFolderPath, "concept_scheme.nt")
		// ### Copy original file to 'processed' sub-folder and work on that from now on
		_, err = copyFile(filepath.Join(versionConfig.SkosSourceFolderPath, "concept_scheme.nt"), version.WorkingFilePathNTriples)
		if err != nil {
			zapLogger.Error(err.Error())
			return err
		}
		// ### Open file and read triples into triple slice
		var tripleDecoder rdf.TripleDecoder
		skosFileReader, err := os.Open(version.WorkingFilePathNTriples)
		defer skosFileReader.Close()
		if err != nil {
			zapLogger.Error(err.Error())
			return err
		}
		var triples []rdf.Triple
		tripleDecoder = rdf.NewTripleDecoder(skosFileReader, rdf.NTriples)
		triples, err = tripleDecoder.DecodeAll()
		if err != nil {
			zapLogger.Error(err.Error())
			return err
		}
		// ### Check for conceptScheme in triples
		conceptSchemeTriples := getMatchingTriples(triples, "", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "http://www.w3.org/2004/02/skos/core#ConceptScheme")
		if len(conceptSchemeTriples) > 0 {
			conceptSchemeUri := conceptSchemeTriples[0].Subj.String() // if there are more than one conceptScheme this just picks the first. This tool expects only one conceptScheme!
			newTriplesRelatedToConceptScheme := make([]rdf.Triple, 0)
			version.Uri = conceptSchemeUri
			subject, _ := rdf.NewIRI(version.Uri)
			var predicate rdf.Predicate
			var object rdf.Object
			existingTriplesRelatedToConceptScheme := getMatchingTriples(triples, conceptSchemeUri, "", "")

			titlesForConceptScheme := getMatchingTriples(existingTriplesRelatedToConceptScheme, conceptSchemeUri, "http://purl.org/dc/terms/title", "")
			if len(titlesForConceptScheme) == 0 {
				predicate, _ = rdf.NewIRI("http://purl.org/dc/terms/title")
				object, _ = rdf.NewLiteral(version.Title)
				newTriplesRelatedToConceptScheme = append(newTriplesRelatedToConceptScheme, rdf.Triple{Subj: subject, Pred: predicate, Obj: object})
			}

			descriptionForConceptScheme := getMatchingTriples(existingTriplesRelatedToConceptScheme, conceptSchemeUri, "http://purl.org/dc/terms/description", "")
			if len(descriptionForConceptScheme) == 0 {
				predicate, _ = rdf.NewIRI("http://purl.org/dc/terms/description")
				object, _ = rdf.NewLiteral(version.Description)
				newTriplesRelatedToConceptScheme = append(newTriplesRelatedToConceptScheme, rdf.Triple{Subj: subject, Pred: predicate, Obj: object})
			}

			creatorsForConceptScheme := getMatchingTriples(existingTriplesRelatedToConceptScheme, conceptSchemeUri, "http://purl.org/dc/terms/creator", "")
			if len(creatorsForConceptScheme) == 0 {
				predicate, _ = rdf.NewIRI("http://purl.org/dc/terms/creator")
				for _, creator := range version.Creators {
					object, _ = rdf.NewLiteral(creator)
					newTriplesRelatedToConceptScheme = append(newTriplesRelatedToConceptScheme, rdf.Triple{Subj: subject, Pred: predicate, Obj: object})
				}
			}

			contributorsForConceptScheme := getMatchingTriples(existingTriplesRelatedToConceptScheme, conceptSchemeUri, "http://purl.org/dc/terms/contributor", "")
			if len(contributorsForConceptScheme) == 0 {
				predicate, _ = rdf.NewIRI("http://purl.org/dc/terms/contributor")
				for _, contributor := range version.Contributors {
					object, _ = rdf.NewLiteral(contributor)
					newTriplesRelatedToConceptScheme = append(newTriplesRelatedToConceptScheme, rdf.Triple{Subj: subject, Pred: predicate, Obj: object})
				}
			}
			for _, newTriple := range newTriplesRelatedToConceptScheme {
				triples = append(triples, newTriple)
			}
		} else {
			return errors.New("ConceptScheme triple NOT found in triples")
		}
		err = writeTriplesToDisk(triples, version.WorkingFilePathNTriples, rdf.NTriples)
		if err != nil {
			return err
		}
		for _, triple := range triples {
			if triple.Obj.String() == "http://www.w3.org/2004/02/skos/core#Concept" {
				conceptUriString := triple.Subj.String()
				triplesForThisConcept := make([]rdf.Triple, 0)
				for _, triple2 := range triples {
					if triple2.Subj.String() == conceptUriString {
						triplesForThisConcept = append(triplesForThisConcept, triple2)
					}
				}
				concept := new(Concept)
				concept.initialise(conceptUriString, version.Namespace, version.Uri, triplesForThisConcept, version.Updated)
				version.Concepts = append(version.Concepts, concept)
			}
		}
		// ### Sort the concepts by title
		sort.Slice(version.Concepts, func(i, j int) bool {
			return version.Concepts[i].Title < version.Concepts[j].Title
		})
		version.InitName(&version, version.ID)
		// ### Configure tree structure
		for _, parentNode := range version.Concepts {
			parentNode.InitName(parentNode, parentNode.ID) //setting the tree id
			for _, childNode := range version.Concepts {
				childNode.InitName(childNode, childNode.ID)
				if contains(childNode.BroaderConcepts, parentNode.ID) {
					parentNode.AddChild(childNode)
				}
			}
			if parentNode.IsTopConcept == true {
				version.AddChild(parentNode)
			}
		}
		conceptScheme.Versions = append(conceptScheme.Versions, version)
		zapLogger.Info(fmt.Sprintf("Initialised concept scheme version: '%s: %s'", config.ID, version.ID))
	}
	sort.Sort(ByVersion(conceptScheme.Versions))
	return err
}

func (conceptScheme *ConceptScheme) GetLatestVersion() ConceptSchemeVersion {
	latestVersion := conceptScheme.Versions[len(conceptScheme.Versions)-1]
	return latestVersion
}
