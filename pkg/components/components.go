package components

// Import catalog component modules
import (
	"github.com/trustacks/catalog/pkg/catalog"
	"github.com/trustacks/catalog/pkg/components/argocd"
	"github.com/trustacks/catalog/pkg/components/authentik"
	"github.com/trustacks/catalog/pkg/components/concourse"
)

func Initialize(catalog *catalog.ComponentCatalog) {
	authentik.Initialize(catalog)
	concourse.Initialize(catalog)
	argocd.Initialize(catalog)
}
