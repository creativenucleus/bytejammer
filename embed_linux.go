package main

import (
	_ "embed"
)

//go:embed build/embed/tic80-linux.exe
var embedTic80exe []byte

//go:embed build/embed/tic80-version.txt
var embedTic80version string
