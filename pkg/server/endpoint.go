package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"liangyuanguo/aw/mr/internal/core"
	"liangyuanguo/aw/mr/pkg/model"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	*core.MrConf
	AuthMethod   model.MrAuthMethod
	TokenAuthUrl string
	BufferSize   int
}

type Endpoint struct {
	Conf           *Config
	ServerUpgrader *websocket.Upgrader

	Con2User map[*websocket.Conn]int64
	User2Con map[int64]*websocket.Conn

	Core *core.MrCore
	quit chan struct{}
}

func (e *Endpoint) GetUserInfo(r *http.Request) (*model.User, error) {
	// TODO: 验证请求来源

	fmt.Println(r.Header)
	if e.Conf.AuthMethod == model.AuthMethodNone {
		devUids, ok := r.Header[model.AuthNoneHeaderUid]
		if !ok || len(devUids) == 0 {
			return nil, fmt.Errorf("no token")
		}
		uid, err := strconv.ParseInt(devUids[0], 10, 64)

		if uid == 0 || err != nil {
			return nil, err
		}

		return &model.User{
			Uid: uid,
		}, nil
	}

	// 检查token
	tokens, ok := r.Header["Authorization"]
	if !ok || len(tokens) == 0 {
		return nil, fmt.Errorf("no token")
	}

	uid, err := e.RequestUserInfo(tokens[0], r.Context())
	if err != nil {
		return nil, err
	}
	return &model.User{
		Uid: uid,
	}, nil
}

func (e *Endpoint) RequestUserInfo(token string, ctx context.Context) (int64, error) {

	//// 创建 OIDC 提供者实例
	//provider, err := oidc.NewProvider(ctx, e.Conf.AuthServerURL+"/realms/"+e.Conf.Realm)
	//if err != nil {
	//	log.Fatalf("Failed to create provider: %v", err)
	//}
	//
	//// 配置 OAuth2 客户端
	//
	//idToken, err := provider.Verifier(&oidc.Config{ClientID: e.Conf.ClientID}).Verify(ctx, token)
	//if err != nil {
	//	return 0, err
	//}
	//var claims struct {
	//	Sub string `json:"sub"`
	//	//Name              string `json:"name"`
	//	//GivenName         string `json:"given_name"`
	//	//FamilyName        string `json:"family_name"`
	//	//Email             string `json:"email"`
	//	//EmailVerified     bool   `json:"email_verified"`
	//	//Picture           string `json:"picture"`
	//	//PreferredUsername string `json:"preferred_username"`
	//	//Profile           string `json:"profile"`
	//	//ZoneInfo          string `json:"zoneinfo"`
	//	//Locale            string `json:"locale"`
	//	//UpdatedAt         int64  `json:"updated_at"`
	//	Exp int64 `json:"exp"`
	//}
	//if err := idToken.Claims(&claims); err != nil {
	//	return 0, err
	//}
	//
	//return strconv.ParseInt(claims.Sub, 10, 64)
	return model.SuperUserId, nil
}

func (e *Endpoint) OnClientConnected(r *http.Request, conn *websocket.Conn) error {
	result, err := e.GetUserInfo(r)
	if err != nil {
		_ = conn.Close()
		return err
	}

	fmt.Println(e.Core.Room[model.DefaultRoomId].Users, result.Uid)
	if u, ok := e.Core.Room[model.DefaultRoomId].Users[result.Uid]; ok {
		// reconnect
		{
			e.Con2User[conn] = result.Uid
			e.User2Con[result.Uid] = conn
		}

		u.IsDown = false

		if rs, ok := e.Core.User2Rooms[u.Uid]; ok {
			for r, _ := range rs {
				_, ok := e.Core.Room[r]
				if !ok {
					continue
				}
				e.Core.RevQ <- model.NewResUserUp(r, u.Uid)
			}
		}

	} else {
		{
			e.Con2User[conn] = result.Uid
			e.User2Con[result.Uid] = conn
		}

		msg := model.NewReqJoinRoom(model.DefaultRoomId, "")
		msg.From = result.Uid
		e.Core.RevQ <- msg
	}

	return nil
}

func (e *Endpoint) OnClientDisconnected(conn *websocket.Conn) {
	userId, ok := e.Con2User[conn]
	if !ok {
		e.Core.Log.Info("OnClientDisconnected: user not found")
		return
	}

	delete(e.Con2User, conn)
	delete(e.User2Con, userId)

	user, ok := e.Core.Room[1].Users[userId]
	if !ok {
		e.Core.Log.Error("OnClientDisconnected: user not found")
		return
	}
	user.IsDown = true
	user.RestKickDuration = e.Core.Conf.ClearInterval
	user.LastDownTime = time.Now()

	if rs, ok := e.Core.User2Rooms[user.Uid]; ok {
		for r, _ := range rs {
			_, ok := e.Core.Room[r]
			if !ok {
				continue
			}
			e.Core.RevQ <- model.NewResUserDown(r, user.Uid)
		}
	}
}

func (e *Endpoint) OnClientMessage(conn *websocket.Conn, msg *model.Message) {
	userId, ok := e.Con2User[conn]
	if !ok {
		_ = fmt.Errorf("OnClientMessage: user not found")
		return
	}

	msg.From = userId
	e.Core.RevQ <- msg
}

func (e *Endpoint) HandleFunc(w http.ResponseWriter, r *http.Request) {
	conn, err := e.ServerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	err = e.OnClientConnected(r, conn)

	defer func() {
		e.OnClientDisconnected(conn)
		err := conn.Close()
		if err != nil {
			return
		}
	}()

	if err != nil {
		e.Core.Log.Error(err.Error())
		return
	}
	for {
		var p []byte
		var msg model.Message
		_, p, err = conn.ReadMessage()
		if err != nil {
			break
		}

		err = json.Unmarshal(p, &msg)
		if err != nil {
			panic(err)
		}
		e.OnClientMessage(conn, &msg)
	}
}

func (e *Endpoint) respond(uid int64, msg *model.Message) {
	a3434, _ := json.Marshal(msg)
	con, ok := e.User2Con[uid]
	if !ok {
		return
	}

	_ = con.WriteMessage(websocket.BinaryMessage, a3434)
}

func (e *Endpoint) Start() {
	e.Core.Start()
	go func() {
		for {
			select {
			case msg := <-e.Core.SendQ:
				if msg.From == model.SuperUserId {
					e.Core.Log.Info("resCmd", "roomId", msg.Room, "to", msg.To, "cmd", msg.Meta["cmd"], "data", msg.Msg)
				} else {
					e.Core.Log.Info("response", "roomId", msg.Room, "to", msg.To, "data", msg.Msg)
				}

				for _, uid := range msg.To {
					if uid == model.SuperUserId {
						continue
					}
					e.respond(uid, msg)
				}
			case <-e.quit:
				break
			}
		}
	}()
}

func (e *Endpoint) Stop() {
	e.Core.Stop()
	e.quit <- struct{}{}
}

func NewServerEndpoint(conf *Config, logger *slog.Logger) *Endpoint {
	mc := core.NewMrCore(conf.MrConf)
	se := &Endpoint{
		Conf: conf,
		ServerUpgrader: &websocket.Upgrader{
			ReadBufferSize:  conf.BufferSize,
			WriteBufferSize: conf.BufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Con2User: make(map[*websocket.Conn]int64),
		User2Con: make(map[int64]*websocket.Conn),
		Core:     mc,
	}

	if logger == nil {
		mc.Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	return se
}
