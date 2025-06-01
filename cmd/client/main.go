package main

import (
	"flag"
	sdk "liangyuanguo/aw/mr/pkg/client"
	api "liangyuanguo/aw/mr/pkg/model"
	"liangyuanguo/aw/mr/pkg/terminal"
)

func main() {
	// 定义命令行参数

	url_ := flag.String("url", "ws://127.0.0.1:8848/", "服务器 URL")
	uid := flag.Int64("uid", 2, "用户 ID")

	// 解析命令行参数
	flag.Parse()

	c := sdk.NewClientEndPoint(*url_, &api.User{Uid: *uid})
	c.Start()
	terminal.GetTerminalInput(c, 1)
	c.Stop()

}
