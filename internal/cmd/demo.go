package cmd

import (
	"fmt"

	feishuLibrary "github.com/armylong/go-library/service/feishu"

	"github.com/urfave/cli/v2"
)

// demo示例命令
func DemoHandler(c *cli.Context) error {
	username := "匿名用户"
	if c.NArg() > 0 {
		username = c.Args().Get(0)
	}
	sex := c.String("sex")

	fmt.Printf("用户：%s\n", username)
	fmt.Printf("性别：%s\n", sex)

	return nil
}

// hello示例命令
func HelloHandler(c *cli.Context) error {
	workHome := c.String("work_home")
	fmt.Printf("工作目录：%s\n", workHome)
	return nil
}

func FeishuHandler(c *cli.Context) error {
	code := "6yCpDbLc98054wee9JHfefyA382B9C7D"
	redirectUri := "https://www.baidu.com"
	userAccessTokenHeader := feishuLibrary.GetUserAccessTokenHeader(&feishuLibrary.GetUserAccessTokenRequest{
		Code:        code,
		RedirectURI: redirectUri,
	})
	fmt.Println(userAccessTokenHeader)
	return nil
}
