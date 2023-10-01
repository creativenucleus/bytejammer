package embed

import (
	_ "embed"
)

//go:embed tic-exe/tic80-macos
var Tic80exe []byte

//go:embed tic-exe/tic80-version.txt
var Tic80version string
