module github.com/AmadlaOrg/garbage

go 1.24.0

replace github.com/AmadlaOrg/LibraryUtils => ../LibraryUtils

replace github.com/AmadlaOrg/LibraryFramework => ../LibraryFramework

require (
	github.com/AmadlaOrg/LibraryFramework v0.0.0
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/olekukonko/tablewriter v0.0.5
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
