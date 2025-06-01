package test

import (
	"fmt"
	"liangyuanguo/aw/mr/internal/core"
	sdk2 "liangyuanguo/aw/mr/pkg/client"
	api "liangyuanguo/aw/mr/pkg/model"
	sdk "liangyuanguo/aw/mr/pkg/server"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
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

	u, _ := url.Parse("ws://127.0.0.1:8848/")
	http.HandleFunc(u.Path, server.HandleFunc)
	server.Start()
	err := http.ListenAndServe(u.Host, nil)
	server.Stop()
	if err != nil {
		return
	}
}

func TestChat(t *testing.T) {

	count := 4
	okC := make(chan struct{})
	for i := range count {

		go func() {
			c := sdk2.NewClientEndPoint("ws://127.0.0.1:8848/", &api.User{Uid: int64(i + 2)})
			c.Start()

			for j := range 10 {
				c.SendQ <- &api.Message{
					Room: 1,
					From: int64(i + 2),
					Msg:  "hello, i'm " + fmt.Sprintf("%d", i) + " " + fmt.Sprintf("%d", j),
					Meta: map[string]any{
						"test": "test",
					},
				}
				time.Sleep(1 * time.Second)
			}

			okC <- struct{}{}
		}()
	}

	for {
		select {
		case <-okC:
			count -= 1
			if count == 0 {
				break
			}
		}
	}

}
