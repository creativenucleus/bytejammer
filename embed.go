package main

import (
	_ "embed"
)

//go:embed build/embed/server.html
var serverHtml []byte

//go:embed build/embed/index.html
var indexHtml []byte

//go:embed build/embed/welcome.lua
var luaWelcome []byte
