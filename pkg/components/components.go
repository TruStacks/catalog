package components

// Import catalog component modules
import (
	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/components/concourse"
	"github.com/trustacks/catalog/pkg/components/flux2"
	"github.com/trustacks/catalog/pkg/components/minio"
	"github.com/trustacks/catalog/pkg/components/sealedsecrets"
	"github.com/trustacks/catalog/pkg/components/vault"
)

func Initialize(catalog *catalog.ComponentCatalog) {
	concourse.Initialize(catalog)
	flux2.Initialize(catalog)
	minio.Initialize(catalog)
	sealedsecrets.Initialize(catalog)
	vault.Initialize(catalog)
}
