package main

import (
	_ "embed"
)

//go:embed build/embed/web/operator.html
var operatorHtml []byte

//go:embed build/embed/web/index.html
var indexHtml []byte

//go:embed build/embed/tic-code/welcome.lua
var luaWelcome []byte

//go:embed build/embed/tic-code/client.lua
var luaClient []byte

//go:embed build/embed/tic-code/author-shim.lua
var luaAuthorShim []byte
