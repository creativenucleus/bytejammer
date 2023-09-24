package main

import (
	_ "embed"
)

// #TODO: FIX!
//
//go:embed build/embed/tic80-win.exe
var embedTic80exe []byte

//go:embed build/embed/tic80-version.txt
var embedTic80version string
