package langserver

import (
	"context"
	"fmt"
	"log"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/tflint"
)

func (h *handler) workspaceDidChangeWatchedFiles(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	if h.rootDir == "" {
		return nil, fmt.Errorf("root directory is undefined")
	}

	newConfig, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, h.configPath)
	if err != nil {
		return nil, err
	}
	newConfig.Merge(h.cliConfig)
	h.config = newConfig

	h.fs = afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	diagnostics, err := h.inspect()
	if err != nil {
		return nil, err
	}

	log.Printf("Notify textDocument/publishDiagnostics with %#v", diagnostics)
	for path, diags := range diagnostics {
		err = conn.Notify(
			ctx,
			"textDocument/publishDiagnostics",
			lsp.PublishDiagnosticsParams{
				URI:         pathToURI(path),
				Diagnostics: diags,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("Failed to notify textDocument/publishDiagnostics: %s", err)
		}
	}

	return nil, nil
}
