// This is a nested module so that the avo code-generation dependency
// never enters the main github.com/woozymasta/png module.
// The generated assembly (../filters_amd64.s) and stubs (../filters_stub_amd64.go)
// are committed, so building or importing the png package never needs avo.
module github.com/woozymasta/png/internal/simd/asmgen

go 1.25.0

require github.com/mmcloughlin/avo v0.6.0

require (
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
)
