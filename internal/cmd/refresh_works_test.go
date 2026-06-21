package cmd

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestRefreshWorksHandler(t *testing.T) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("work_home", "/Users/zhangzelong/works/rubrics", "work home")
	cliCtx := cli.NewContext(nil, flagSet, nil)
	RefreshWorksHandler(cliCtx)
}
