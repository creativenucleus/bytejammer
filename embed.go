package main

import (
	_ "embed"
)

//go:embed build/embed/operator.html
var operatorHtml []byte

//go:embed build/embed/index.html
var indexHtml []byte

//go:embed build/embed/welcome.lua
var luaWelcome []byte

//go:embed build/embed/client.lua
var luaClient []byte
