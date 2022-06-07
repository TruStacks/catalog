package components

// Import catalog component modules
import (
	_ "github.com/trustacks/catalog/pkg/components/concourse"
	_ "github.com/trustacks/catalog/pkg/components/flux2"
	_ "github.com/trustacks/catalog/pkg/components/minio"
	_ "github.com/trustacks/catalog/pkg/components/vault"
)

var hooks = make(map[string]func() error)
