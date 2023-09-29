package main

import (
	_ "embed"
)

//go:embed build/embed/web/server/operator.html
var serverOperatorHtml []byte

//go:embed build/embed/web/server/index.html
var serverIndexHtml []byte

//go:embed build/embed/web/client/index.html
var clientIndexHtml []byte

//go:embed build/embed/tic-code/welcome.lua
var luaWelcome []byte

//go:embed build/embed/tic-code/client.lua
var luaClient []byte

//go:embed build/embed/tic-code/author-shim.lua
var luaAuthorShim []byte
