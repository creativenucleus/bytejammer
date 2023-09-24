package main

import (
	_ "embed"
)

//go:embed build/embed/tic80-macos
var embedTic80exe []byte

//go:embed build/embed/tic80-version.txt
var embedTic80version string
