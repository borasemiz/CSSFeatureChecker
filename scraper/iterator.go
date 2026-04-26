package scraper

import (
	"encoding/json"
	"errors"
	"io"
)

type FeatureInterator interface {
	Next() (*Feature, error)
}

type featureIteratorFromJson struct {
	decoder             *json.Decoder
	currentNestingLevel int
	isParsingDataObject bool
}

func (it *featureIteratorFromJson) Next() (*Feature, error) {
	decoder := it.decoder
	var feature Feature

	if !it.isParsingDataObject {
		if err := it.skipToFeatures(); err != nil {
			return nil, err
		}
	}

	v, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	val, ok := v.(string)
	if !ok {
		return nil, errors.New("end_features")
	}

	feature.ID = val
	if err := decoder.Decode(&feature); err != nil {
		return nil, err
	}

	return &feature, nil
}

func (it *featureIteratorFromJson) skipToFeatures() error {
	decoder := it.decoder

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		if v, ok := token.(json.Delim); ok {
			if v == '{' {
				it.currentNestingLevel += 1

				if it.isParsingDataObject {
					break
				}
			}

			if v == '}' {
				it.currentNestingLevel -= 1
			}
		}

		if v, ok := token.(string); ok {
			if it.currentNestingLevel == 1 && v == "data" {
				it.isParsingDataObject = true
			}
		}
	}

	return nil
}

func MakeFeatureIteratorFromJSON(jsonReader io.Reader) FeatureInterator {
	return &featureIteratorFromJson{
		decoder: json.NewDecoder(jsonReader),
	}
}
