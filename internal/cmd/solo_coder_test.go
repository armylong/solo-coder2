package cmd

import (
	"context"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestSoloCoderSoloCoderUploadFeishuHandler(t *testing.T) {
	cliCtx := &cli.Context{
		Context: context.Background(),
	}
	SoloCoderUploadFeishuHandler(cliCtx)
}
