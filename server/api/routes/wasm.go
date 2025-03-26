package routes

import (
	_ "embed"
	"net/http"

	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

//go:generate cp /usr/local/go/misc/wasm/wasm_exec.js wasm_exec.js
//go:embed wasm_exec.js
var wasmExec []byte

func (h *Handler) GetWasm(c gateway.Context) error {
	return c.Blob(http.StatusOK, "text/javascript", wasmExec)
}
