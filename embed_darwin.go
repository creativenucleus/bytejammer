package main

import (
	_ "embed"
)

// #TODO: UPDATE

//go:embed embed/tic80-macos.exe
var embedTic80exe []byte

//go:embed embed/tic80-version.txt
var embedTic80version string
