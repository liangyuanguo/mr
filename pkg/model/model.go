package model

import (
	"time"
)

type MrAuthMethod string

const AuthNoneHeaderUid = "Uid"

const (
	// AuthMethodNone 你说你是谁就是谁
	AuthMethodNone MrAuthMethod = "none"
	// AuthMethodBasic 最好不要用这个
	AuthMethodBasic MrAuthMethod = "basic"
	AuthMethodToken MrAuthMethod = "token"
)

const (
	// SuperUserId 也是系统管理员
	SuperUserId   int64 = 1
	DefaultGroup        = 1
	DefaultRoomId       = 1
)

type CmdType int

type User struct {
	Uid int64 `json:"uid"`

	IsSuperUser      bool          `json:"isSuperUser"`
	IsDown           bool          `json:"isDown"`
	LastDownTime     time.Time     `json:"lastDownTime"`
	RestKickDuration time.Duration `json:"restKickDuration"`
}

type Room struct {
	ID    int64           `json:"id"`
	Name  string          `json:"name"`
	Users map[int64]*User `json:"users"`
	Owner *User           `json:"owner"`

	MaxUserCnt int
	Password   string `json:"-"`
	IsPublic   bool   `json:"isPublic"` // searchable if true
	// 最大255字符
	Desc string `json:"desc"`
}

type Message struct {
	Room int64 `json:"room"`
	From int64 `json:"from"`
	// 1, 管理员, 0, 所有人
	Meta map[string]any `json:"meta"`
	To   []int64        `json:"to"`
	Msg  any            `json:"msg"`
}
