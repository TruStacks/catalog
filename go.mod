module github.com/trustacks/catalog

go 1.18

replace (
	github.com/trustacks/catalog/pkg => ./pkg
	github.com/trustacks/catalog/server => ./server
)

require github.com/trustacks/catalog/pkg v0.0.0-00010101000000-000000000000
