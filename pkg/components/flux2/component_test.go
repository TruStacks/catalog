package flux2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/trustacks/catalog/pkg/catalog"
	"gopkg.in/yaml.v3"
)

func TestLoadComponent(t *testing.T) {
	var config catalog.BaseComponent
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
}

func TestGetChart(t *testing.T) {
	var config catalog.BaseComponent
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/%s-%s.tgz", config.Repo, config.Chart, config.Version)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("failed retrieving the helm chart")
	}
}
