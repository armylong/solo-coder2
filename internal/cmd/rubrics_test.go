package cmd

import (
	"context"
	"flag"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestRubricsHandler(t *testing.T) {
	cliCtx := &cli.Context{
		Context: context.Background(),
	}
	RubricsHandler(cliCtx)
}

func TestRubricsDownloadHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("work_home", "/Users/zhangzelong/works/rubrics", "work home")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	RubricsDownloadHandler(cliCtx)
}

func TestRubricsUploadHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("work_home", "/Users/zhangzelong/works/rubrics", "work home")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	RubricsUploadHandler(cliCtx)
}
