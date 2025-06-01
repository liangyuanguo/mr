package core

import (
	"fmt"
	"liangyuanguo/aw/mr/internal/util"
	"liangyuanguo/aw/mr/pkg/model"
	"log/slog"
	"os"
	"time"
)

type MrConf struct {
	MaxChanLen int

	MaxUserCnt int
	MaxRoomCnt int

	MaxRoomPerNormalUser int
	MaxUserPerRoom       int

	// 管理员密码
	SuSecret string

	// 数据清理间隔
	ClearInterval time.Duration
}

type MrCore struct {
	Conf *MrConf

	Room       map[int64]*model.Room
	User2Rooms map[int64]map[int64]bool

	SendQ chan *model.Message
	RevQ  chan *model.Message

	Ticker *time.Ticker
	Quit   chan struct{}
	Log    *slog.Logger
}

func NewMrCore(conf *MrConf) *MrCore {
	mc := &MrCore{
		Room:       make(map[int64]*model.Room),
		User2Rooms: make(map[int64]map[int64]bool),
		SendQ:      make(chan *model.Message, conf.MaxChanLen),
		RevQ:       make(chan *model.Message, conf.MaxChanLen),
		Conf:       conf,
		Ticker:     time.NewTicker(conf.ClearInterval),
		Quit:       make(chan struct{}),
	}

	mc.Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn, // 设置日志级别为 Info
	}))

	root := mc.InitUser(model.SuperUserId)
	root.IsSuperUser = true
	publicRoom := utility.NewRoom("global", root)
	publicRoom.ID = model.SuperUserId
	mc.addRoom(publicRoom)
	return mc
}

func (e *MrCore) InitUser(uid int64) *model.User {
	return &model.User{
		Uid:              uid,
		IsDown:           false,
		LastDownTime:     time.Now(),
		RestKickDuration: e.Conf.ClearInterval,
	}
}

// Route
// From 是已经处理好的，User已经创建好
func (e *MrCore) Route(msg *model.Message) {
	// 命令消息
	if msg.To != nil && len(msg.To) > 0 && msg.To[0] == model.SuperUserId { // check is system
		{
			if msg.Meta == nil {
				e.Log.Warn("msg is nil")
				return
			}
			if _, ok := msg.Meta["cmd"]; !ok {
				e.Log.Warn("cmd not found")
				return
			}
		}

		var user *model.User
		if _, ok := e.Room[1].Users[msg.From]; !ok {
			user = e.InitUser(msg.From)
		} else {
			user = e.Room[1].Users[msg.From]
		}

		// 写需要阻塞
		err := e.OnSystemCmd(user, msg)
		if err != nil {
			e.Log.Error(err.Error())
			return
		}
		return
	}

	room, ok := e.Room[msg.Room]
	if !ok {
		e.Log.Warn("room not found")
		return
	}

	toUidArr := make([]int64, 0)
	if msg.To == nil || len(msg.To) == 0 {
		groupUsers := room.Users
		for user2id, _ := range groupUsers {
			if user2id == model.SuperUserId {
				continue
			}
			toUidArr = append(toUidArr, user2id)
		}
	} else {
		for _, user2id := range msg.To {
			if user2id == model.SuperUserId {
				continue
			}
			if _, ok := room.Users[user2id]; ok {
				toUidArr = append(toUidArr, user2id)
			}
		}
	}
	msg.To = toUidArr

	if len(toUidArr) > 0 {
		e.Log.Debug("req", "roomID", msg.Room, "from", msg.From, "to", fmt.Sprintf("%v", msg.To), "data", msg.Msg)
		e.SendQ <- msg
	}
}

func (e *MrCore) Start() {
	e.Ticker = time.NewTicker(time.Second * 30)

	go func() {
		for {
			select {
			case <-e.Ticker.C:
				e.RevQ <- model.NewReqClear()
			case <-e.Quit:
				e.Ticker.Stop()
				return
			case msg := <-e.RevQ:
				e.Route(msg)
			}
		}

	}()
}

func (e *MrCore) Stop() {
	e.Quit <- struct{}{}
	e.Quit <- struct{}{}
}
