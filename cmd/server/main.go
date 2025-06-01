package main

import (
	"flag"
	"fmt"
	"liangyuanguo/aw/mr/internal/core"
	api "liangyuanguo/aw/mr/pkg/model"
	sdk "liangyuanguo/aw/mr/pkg/server"
	"net/http"
	"net/url"
	"time"
)

func main() {
	// 定义命令行参数

	url_ := flag.String("url", "ws://127.0.0.1:8848/", "服务器 URL")
	// 解析命令行参数
	flag.Parse()

	// 创建配置结构体实例
	config := &sdk.Config{
		AuthMethod: api.AuthMethodNone,
		BufferSize: 2048,
		MrConf: &core.MrConf{
			MaxChanLen:           100,
			MaxRoomCnt:           100,
			MaxRoomPerNormalUser: 10,
			MaxUserCnt:           100,
			MaxUserPerRoom:       10,
			SuSecret:             "123456",
			ClearInterval:        30 * time.Second,
		},
	}

	// 打印配置信息
	fmt.Printf(" Config: %+v\n", config)

	server := sdk.NewServerEndpoint(config, nil)

	u, _ := url.Parse(*url_)
	http.HandleFunc(u.Path, server.HandleFunc)
	server.Start()
	err := http.ListenAndServe(u.Host, nil)
	server.Stop()
	if err != nil {
		return
	}

}
