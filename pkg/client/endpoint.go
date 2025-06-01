package client

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	api "liangyuanguo/aw/mr/pkg/model"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

type EndPoint struct {
	Url string

	SendQ chan *api.Message
	RevQ  chan *api.Message

	Room map[int64]*api.Room

	User  *api.User `json:"user"`
	Token string

	Conn *websocket.Conn
	Quit chan struct{}
	Log  *slog.Logger

	// 最大重试次数
	MaxRetryCount int
	// 重试间隔
	RetryInterval time.Duration
}

func (e *EndPoint) OnServerConnected(conn *websocket.Conn) {
	e.Conn = conn
	e.User.IsDown = false
}

func (e *EndPoint) OnServerDisconnected(conn *websocket.Conn) {
	e.User.IsDown = true
	e.Quit <- struct{}{}
}

func (e *EndPoint) tryConnect() bool {
	isOk := false
	for i := 0; i < e.MaxRetryCount; i += 1 {
		time.Sleep(e.RetryInterval)
		c, response, err := websocket.DefaultDialer.Dial(e.Url, http.Header{
			api.AuthNoneHeaderUid: []string{strconv.Itoa(int(e.User.Uid))},
		})
		if err != nil {
			e.Log.Debug("重连失败", "error", err)
			continue
		}
		e.Log.Debug("重连成功", "status", response.Status)
		e.OnServerConnected(c)
		isOk = true
		break
	}

	if !isOk {
		return false
	}
	return true
}

func (e *EndPoint) Start() {
	e.Log.Info("userId", "uid", e.User.Uid)
	if !e.tryConnect() {
		e.Log.Error("连接失败")
		e.Quit <- struct{}{}
	}
	go func() {
		defer func() {
			e.OnServerDisconnected(e.Conn)
		}()
		for {
			select {
			case <-e.Quit:
				return
			default:
				var p []byte
				var msg api.Message
				_, p, err := e.Conn.ReadMessage()
				if err != nil {
					if !e.tryConnect() {
						e.Log.Error("连接失败")
						e.Quit <- struct{}{}
					}
				}

				err = json.Unmarshal(p, &msg)
				if err != nil {
					e.Log.Error("解析消息失败", "message", p, "error", err)
					continue
				}
				if msg.From == api.SuperUserId {
					e.Log.Debug("resCmd", "cmd", msg.Meta["cmd"], "content", msg.Msg)
				} else {
					e.Log.Debug("res", "content", msg.Msg)
				}
				e.RevQ <- &msg
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-e.SendQ:
				if e.User.IsDown {
					time.Sleep(e.RetryInterval)
					continue
				}
				p, err := json.Marshal(msg)
				if err != nil {
					panic(err)
				}

				if msg.To != nil && len(msg.To) == 1 && msg.To[0] == api.SuperUserId {
					e.Log.Debug("reqCmd", "cmd", msg.Meta["cmd"], "content", msg.Msg)
				} else {
					e.Log.Debug("req", "content", msg.Msg)
				}

				err = e.Conn.WriteMessage(websocket.BinaryMessage, p)
				if err != nil {
					return
				}
			case <-e.Quit:
				return
			}
		}
	}()
}

func (e *EndPoint) Stop() {
	close(e.Quit)
	err := e.Conn.Close()
	if err != nil {
		return
	}
}

func NewClientEndPoint(url string, user *api.User) *EndPoint {
	println(user.Uid)
	return &EndPoint{
		Url:   url,
		SendQ: make(chan *api.Message, 128),
		RevQ:  make(chan *api.Message, 128),
		Room:  make(map[int64]*api.Room),
		User:  user,
		Quit:  make(chan struct{}),
		Log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // 设置日志级别为 Info
		})),
		MaxRetryCount: 12,
		RetryInterval: 5 * time.Second,
	}
}
