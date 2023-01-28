package convert

import (
	"encoding/json"
	"errors"

	"github.com/practable/book/internal/client/models"
	"github.com/practable/book/internal/store"
	"sigs.k8s.io/yaml"
)

func JSONToManifests(jb []byte) (models.Manifest, store.Manifest, error) {

	m := models.Manifest{}
	s := store.Manifest{}

	err := json.Unmarshal(jb, &s)
	if err != nil {
		return m, s, errors.New("unable to unmarshal manifest into store format because " + err.Error())
	}

	err = json.Unmarshal(jb, &m)

	if err != nil {
		return m, s, errors.New("unable to unmarshal manifest into client format because " + err.Error())
	}

	return m, s, nil
}

func YAMLToManifests(yb []byte) (models.Manifest, store.Manifest, error) {

	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		return models.Manifest{}, store.Manifest{}, errors.New("unable to process manifest because " + err.Error())
	}

	return JSONToManifests(jb)

}
