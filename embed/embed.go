package embed

import (
	_ "embed"
)

//go:embed web/server/operator.html
var ServerOperatorHtml []byte

//go:embed web/server/index.html
var ServerIndexHtml []byte

//go:embed web/client/index.html
var ClientIndexHtml []byte

//go:embed tic-code/jukebox.lua
var LuaJukebox []byte

//go:embed tic-code/client.lua
var LuaClient []byte

//go:embed tic-code/author-shim.lua
var LuaAuthorShim []byte
