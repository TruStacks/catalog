package concourse

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/trustacks/catalog/pkg/catalog"
	"gopkg.in/yaml.v3"
)

func TestGetChart(t *testing.T) {
	var conf *catalog.ComponentConfig
	if err := yaml.Unmarshal(config, &conf); err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/%s-%s.tgz", conf.Repo, conf.Chart, conf.Version)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("failed retrieving the helm chart")
	}
}
