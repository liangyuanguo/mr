package model

const (
	ReqRooms       = "reqRooms"
	ResRooms       = "resRooms"
	ReqRoom        = "reqRoom"
	ResRoom        = "resRoom"
	ReqAddRoom     = "reqAddRoom"
	ReqDelRoom     = "reqDelRoom"
	ResRoomDeleted = "resRoomDeleted"
	ReqSetRoom     = "reqSetRoom"
	ReqJoin        = "reqJoin"
	ReqQuit        = "reqQuit"
	ReqKick        = "reqKick"
	ResUserUp      = "resUserUp"
	ResUserDown    = "resUserDown"
	ResUserJoined  = "resUserJoined"
	ResUserQuit    = "resUserQuit"
	ResUserKicked  = "resUserKicked"
	ReqSu          = "reqSu"

	ReqNodeInfo = "reqNodeInfo"
	ResNodeInfo = "resNodeInfo"
	ReqClear    = "reqClear"
)

func NewReqRooms() *Message {
	return &Message{
		Room: DefaultRoomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqRooms,
		},
	}
}

type CmdResRooms []struct {
	Id      int64
	Name    string
	Desc    string
	OwnerId int64
}

func NewResRooms(to []int64, res CmdResRooms) *Message {
	return &Message{
		From: SuperUserId,
		Room: DefaultRoomId,
		To:   to,
		Msg:  res,
		Meta: map[string]any{
			"cmd": ResRooms,
		},
	}
}

func NewReqRoom(roomId int64) *Message {
	return &Message{
		Room: roomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqRoom,
		},
	}
}

type CmdResRoom struct {
	Id      int64
	Name    string
	Desc    string
	OwnerId int64

	Users map[int64]struct{}

	UserCnt    int
	MaxUserCnt int
	IsPublic   bool
	Groups     map[int64]map[int64]struct{}
}

func NewResRoom(to []int64, res *CmdResRoom) *Message {
	return &Message{
		Room: DefaultRoomId,
		From: SuperUserId,
		To:   to,
		Msg:  res,
		Meta: map[string]any{
			"cmd": ResRoom,
		},
	}
}

type CmdReqAddRoom struct {
	Name       string
	Desc       string
	MaxUserCnt int
	IsPublic   bool
	Password   string
	Groups     map[int64]map[int64]struct{}
}

func NewReqAddRoom(req *CmdReqAddRoom) *Message {
	return &Message{
		Room: DefaultRoomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqAddRoom,
		},
		Msg: req,
	}
}

func NewReqDelRoom(roomId int64) *Message {
	return &Message{
		Room: DefaultRoomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqDelRoom,
		},
		Msg: roomId,
	}
}

func NewResRoomDeleted(roomId int64) *Message {
	return &Message{
		From: SuperUserId,
		Room: DefaultRoomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ResRoomDeleted,
		},
		Msg: roomId,
	}
}

type CmdReqSetRoom struct {
	Name    *string
	Desc    *string
	OwnerId *int64

	MaxUserCnt *int
	IsPublic   *bool
	Password   *string

	Groups map[int64]map[int64]struct{}
}

func NewReqSetRoom(room int64, req *CmdReqSetRoom) *Message {
	return &Message{
		Room: room,
		To:   []int64{SuperUserId},

		Meta: map[string]any{
			"cmd": ReqSetRoom,
		},

		Msg: req,
	}
}

func NewReqJoinRoom(roomId int64, password string) *Message {
	return &Message{
		Room: roomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqJoin,
		},
		Msg: password,
	}
}

func NewReqQuitRoom(roomId int64) *Message {
	return &Message{
		Room: roomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqQuit,
		},
	}
}

func NewReqKickUser(roomId int64, userId int64) *Message {
	return &Message{
		Room: roomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqKick,
		},
		Msg: userId,
	}
}

func NewResUserUp(room int64, userId int64) *Message {
	return &Message{
		From: SuperUserId,
		Room: room,
		Meta: map[string]any{
			"cmd": ResUserUp,
		},
		Msg: userId,
	}
}

func NewResUserDown(room int64, userId int64) *Message {
	return &Message{
		From: SuperUserId,
		Room: room,
		Meta: map[string]any{
			"cmd": ResUserDown,
		},
		Msg: userId,
	}
}

func NewResUserJoined(room int64, userId int64) *Message {
	return &Message{
		Room: room,
		From: SuperUserId,
		Meta: map[string]any{
			"cmd": ResUserJoined,
		},
		Msg: userId,
	}
}

func NewResUserQuit(room int64, userId int64) *Message {
	return &Message{
		Room: room,
		From: SuperUserId,
		Meta: map[string]any{
			"cmd": ResUserQuit,
		},
		Msg: userId,
	}
}

func NewResUserKicked(room int64, userId int64) *Message {
	return &Message{
		From: SuperUserId,
		Room: room,
		Meta: map[string]any{
			"cmd": ResUserKicked,
		},
		Msg: userId,
	}
}

//================================== CMD FOR ADMIN ====================================

func NewReqSu(password string) *Message {
	return &Message{
		Room: DefaultRoomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqSu,
		},
		Msg: password,
	}
}

func NewReqNodeInfo() *Message {
	return &Message{
		Room: DefaultRoomId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqNodeInfo,
		},
	}
}

type CmdResNodeInfo struct {
	RoomCnt int
	UserCnt int
}

func NewResNodeInfo(to []int64, info *CmdResNodeInfo) *Message {
	return &Message{
		From: SuperUserId,
		Room: DefaultRoomId,
		To:   to,
		Meta: map[string]any{
			"cmd": ResNodeInfo,
		},
		Msg: info,
	}
}

func NewReqClear() *Message {
	return &Message{
		Room: DefaultRoomId,
		From: SuperUserId,
		To:   []int64{SuperUserId},
		Meta: map[string]any{
			"cmd": ReqClear,
		},
	}
}

func ParseCmd(cmd *CmdParser) (*Message, error) {
	switch cmd.Args[0] {
	case ReqRoom:
		return NewReqRoom(int64(cmd.GetArgInt(1, 1))), nil
	case ReqRooms:
		return NewReqRooms(), nil
	case ReqAddRoom:
		return NewReqAddRoom(&CmdReqAddRoom{
			Name:       cmd.GetOptStr("name", "name"),
			Desc:       cmd.GetOptStr("desc", "desc"),
			MaxUserCnt: cmd.GetOptInt("maxUserCnt", 16),
			IsPublic:   cmd.GetOptBool("isPublic", false),
			Password:   cmd.GetOptStr("password", ""),
		}), nil
	case ReqSetRoom:
		Name := cmd.GetOptStr("name", "")
		Desc := cmd.GetOptStr("desc", "")
		MaxUserCnt := cmd.GetOptInt("maxUserCnt", 0)
		IsPublic := cmd.GetOptBool("isPublic", false)
		Password := cmd.GetOptStr("password", "")
		return NewReqSetRoom(int64(cmd.GetArgInt(1, 1)), &CmdReqSetRoom{
			Name:       &Name,
			Desc:       &Desc,
			MaxUserCnt: &MaxUserCnt,
			IsPublic:   &IsPublic,
			Password:   &Password,
		}), nil
	case ReqDelRoom:
		return NewReqDelRoom(int64(cmd.GetArgInt(1, 1))), nil
	case ReqJoin:
		return NewReqJoinRoom(int64(cmd.GetArgInt(1, 1)), cmd.Args[2]), nil
	case ReqQuit:
		return NewReqQuitRoom(int64(cmd.GetArgInt(1, 1))), nil
	case ReqKick:
		return NewReqKickUser(int64(cmd.GetArgInt(1, 1)), int64(cmd.GetArgInt(2, 1))), nil
	case ReqSu:
		return NewReqSu(cmd.Args[1]), nil
	case ReqNodeInfo:
		return NewReqNodeInfo(), nil
	default:
		return nil, nil
	}
}
