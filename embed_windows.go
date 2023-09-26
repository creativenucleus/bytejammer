package main

import (
	_ "embed"
)

//go:embed build/embed/tic-exe/tic80-win.exe
var embedTic80exe []byte

//go:embed build/embed/tic-exe/tic80-version.txt
var embedTic80version string
