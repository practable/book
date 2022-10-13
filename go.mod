module github.com/timdrysdale/interval

go 1.18

require (
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.8.0
	internal/trees v1.0.0

)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	internal/containers v0.0.0-00010101000000-000000000000 // indirect
	internal/utils v0.0.0-00010101000000-000000000000 // indirect
)

replace internal/trees => ./internal/trees

replace internal/containers => ./internal/containers

replace internal/utils => ./internal/utils

replace internal/resource => ./internal/resource
