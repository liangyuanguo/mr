package core

import (
	"fmt"
	"liangyuanguo/aw/mr/internal/util"
	"liangyuanguo/aw/mr/pkg/model"
	"math/rand"
	"time"
)

func (e *MrCore) removeRoom(r *model.Room) {
	for k, _ := range r.Users {
		delete(e.User2Rooms[k], r.ID)
	}
	delete(e.Room, r.ID)
}

func (e *MrCore) getResRoom(roomId int64) (*model.CmdResRoom, error) {
	room, ok := e.Room[roomId]
	if !ok {
		return nil, fmt.Errorf("room not found")
	}

	res := &model.CmdResRoom{
		Id:         room.ID,
		Name:       room.Name,
		Desc:       room.Desc,
		OwnerId:    room.Owner.Uid,
		Users:      make(map[int64]struct{}, len(room.Users)),
		IsPublic:   room.IsPublic,
		MaxUserCnt: room.MaxUserCnt,
		UserCnt:    len(room.Users),
	}

	for k, _ := range room.Users {
		res.Users[k] = struct{}{}
	}

	return res, nil

}

func (e *MrCore) addRoom(r *model.Room) {
	id := r.ID
	for {
		if _, ok := e.Room[id]; !ok && id > 0 {
			e.Room[id] = r
			r.ID = id
			break
		}
		id = int64(rand.Uint32())
	}

	if _, ok := e.User2Rooms[r.Owner.Uid]; !ok {
		e.User2Rooms[r.Owner.Uid] = make(map[int64]bool)
	}

	e.User2Rooms[r.Owner.Uid][r.ID] = true
	e.Room[r.ID] = r
}

func (e *MrCore) OnSystemCmd(user *model.User, msg *model.Message) error {
	cmdKey, _ := msg.Meta["cmd"]

	e.Log.Debug("reqCmd", "roomId", msg.Room, "cmd", cmdKey, "data", msg.Msg)
	switch cmdKey {
	case model.ReqRooms:
		return e.HandleReqRooms(msg, user)
	case model.ReqRoom:
		return e.HandleReqRoom(msg, user)
	case model.ReqAddRoom:
		return e.HandleReqAddRoom(msg, user)
	case model.ReqDelRoom:
		return e.HandleReqDelRoom(msg, user)
	case model.ReqSetRoom:
		return e.HandleReqSetRoom(msg, user)
	case model.ReqJoin:
		return e.HandleReqJoin(msg, user)
	case model.ReqKick:
		return e.HandleReqKick(msg, user)
	case model.ReqQuit:
		return e.HandleReqQuit(msg, user)
	case model.ReqSu:
		return e.HandleReqSu(msg, user)
	case model.ReqNodeInfo:
		return e.HandleReqNodeInfo(msg, user)
	case model.ReqClear:
		return e.HandleReqClear(msg, user)
	default:
		return fmt.Errorf("unknown cmdKey")
	}
}

func (e *MrCore) HandleReqRooms(_ *model.Message, user *model.User) error {
	res := model.CmdResRooms{}
	for roomId := range e.User2Rooms[user.Uid] {
		room, ok := e.Room[roomId]
		if !ok {
			e.Log.Error("HandleReqRooms: room not found")
			continue
		}

		res = append(res, struct {
			Id      int64
			Name    string
			Desc    string
			OwnerId int64
		}{
			Id:      roomId,
			Name:    room.Name,
			Desc:    room.Desc,
			OwnerId: room.Owner.Uid,
		})
	}
	resMsg := model.NewResRooms([]int64{user.Uid}, res)

	e.SendQ <- resMsg
	return nil
}

func (e *MrCore) HandleReqRoom(msg *model.Message, user *model.User) error {
	roomId := msg.Room

	room, err := e.getResRoom(roomId)
	if err != nil {
		e.Log.Error("HandleReqRoom: getResRoom failed")
		return err
	}

	resMsg := model.NewResRoom([]int64{user.Uid}, room)

	e.SendQ <- resMsg
	return nil
}

func (e *MrCore) HandleReqAddRoom(msg *model.Message, user *model.User) error {

	req, ok := msg.Msg.(model.CmdReqAddRoom)
	if !ok {
		return fmt.Errorf("invalid req")
	}

	if !user.IsSuperUser {
		if v, ok := e.User2Rooms[user.Uid]; ok && len(v) >= e.Conf.MaxRoomPerNormalUser {
			return fmt.Errorf("too many rooms")
		}
		if len(e.Room) >= e.Conf.MaxRoomCnt {
			return fmt.Errorf("too many rooms")
		}
	}

	room := utility.NewRoom(req.Name, user)
	room.IsPublic = req.IsPublic
	room.MaxUserCnt = min(req.MaxUserCnt, e.Conf.MaxUserPerRoom)
	room.Password = req.Password
	room.Desc = req.Desc
	if len(room.Desc) > 255 {
		room.Desc = room.Desc[:255]
	}

	e.addRoom(room)

	room2, err := e.getResRoom(room.ID)
	if err != nil {
		return err
	}

	resMsg := model.NewResRoom([]int64{user.Uid}, room2)

	e.SendQ <- resMsg
	return nil
}

func (e *MrCore) HandleReqDelRoom(msg *model.Message, user *model.User) error {
	roomId := msg.Msg.(int64)
	if roomId == 0 {
		return fmt.Errorf("invalid roomId")
	}

	room, ok := e.Room[roomId]
	if !ok {
		return fmt.Errorf("room not found")
	}
	if room.Owner.Uid != user.Uid && !user.IsSuperUser {
		return fmt.Errorf("not owner")
	}

	resp := model.NewResRoomDeleted(roomId)

	var toUsers []int64
	for k, _ := range room.Users {
		if k == model.SuperUserId {
			continue
		}
		toUsers = append(toUsers, k)
	}

	resp.To = toUsers
	e.SendQ <- resp

	for k, _ := range room.Users {
		delete(e.User2Rooms[k], room.ID)
	}
	delete(e.Room, room.ID)
	return nil
}
func (e *MrCore) HandleReqSetRoom(msg *model.Message, user *model.User) error {
	if msg.Room == model.DefaultRoomId && !user.IsSuperUser {
		return fmt.Errorf("not allow")
	}

	req := msg.Msg.(model.CmdReqSetRoom)

	room, ok := e.Room[msg.Room]
	if !ok {
		return fmt.Errorf("room not found")
	}
	if room.Owner.Uid != user.Uid && !user.IsSuperUser {
		return fmt.Errorf("not owner")
	}

	if req.Name != nil {
		room.Name = *req.Name
	}

	if req.Desc != nil {
		room.Desc = *req.Desc
		if len(room.Desc) > 255 {
			room.Desc = room.Desc[:255]
		}
	}

	if req.MaxUserCnt != nil {
		room.MaxUserCnt = min(*req.MaxUserCnt, e.Conf.MaxUserPerRoom)
	}

	if req.IsPublic != nil {
		room.IsPublic = *req.IsPublic
	}

	if req.Password != nil {
		room.Password = *req.Password
	}

	if req.OwnerId != nil && msg.Room != model.DefaultRoomId {
		if newOwner, ok := room.Users[*req.OwnerId]; ok {
			room.Owner = newOwner
		} else {
			e.Log.Warn("HandleReqSetRoom: owner not found")
		}
	}

	roomRes, err := e.getResRoom(room.ID)
	if err != nil {
		return err
	}

	resMsg := model.NewResRoom([]int64{user.Uid}, roomRes)
	e.SendQ <- resMsg
	return nil
}

func (e *MrCore) HandleReqJoin(msg *model.Message, user *model.User) error {
	roomId := msg.Room
	room, ok := e.Room[roomId]
	if !ok {
		return fmt.Errorf("room not found")
	}

	if _, ok := room.Users[user.Uid]; !ok {
		if room.Password != msg.Msg {
			return fmt.Errorf("password error")
		}

		if room.ID != model.DefaultRoomId && len(room.Users) >= room.MaxUserCnt {
			return fmt.Errorf("room full")
		}
		utility.AddMem(room, user)
		if _, ok := e.User2Rooms[user.Uid]; !ok {
			e.User2Rooms[user.Uid] = make(map[int64]bool)
		}
		e.User2Rooms[user.Uid][roomId] = true

		uids := make([]int64, 0)
		for k := range room.Users {
			uids = append(uids, k)
		}

		resMsg := model.NewResUserJoined(roomId, user.Uid)
		resMsg.To = uids

		e.SendQ <- resMsg
	}

	roomRes, err := e.getResRoom(roomId)
	if err != nil {
		return err
	}
	e.SendQ <- model.NewResRoom([]int64{user.Uid}, roomRes)
	return nil
}

func (e *MrCore) HandleReqKick(msg *model.Message, user *model.User) error {
	targetUid := msg.Msg.(int64)
	roomId := msg.Room
	room, ok := e.Room[roomId]
	if !ok {
		return fmt.Errorf("room not found")
	}
	if room.Owner.Uid != user.Uid && !user.IsSuperUser {
		return fmt.Errorf("not owner")
	}

	if u, ok := room.Users[targetUid]; ok {
		utility.RemMem(room, u)
		delete(e.User2Rooms[targetUid], roomId)
	} else {
		return fmt.Errorf("user not in room")
	}

	uids := make([]int64, 0)
	for k := range room.Users {
		uids = append(uids, k)
	}
	uids = append(uids, targetUid)
	resMsg := model.NewResUserKicked(roomId, targetUid)
	resMsg.To = uids
	e.SendQ <- resMsg
	return nil
}

func (e *MrCore) HandleReqQuit(msg *model.Message, user *model.User) error {
	roomId := msg.Room
	room, ok := e.Room[roomId]
	if !ok {
		return fmt.Errorf("room not found")
	}
	if u, ok := room.Users[user.Uid]; ok {
		utility.RemMem(room, u)
		delete(e.User2Rooms[user.Uid], roomId)
	} else {
		return fmt.Errorf("user not in room")
	}

	uids := make([]int64, 0)
	for k := range room.Users {
		uids = append(uids, k)
	}
	uids = append(uids, user.Uid)

	resMsg := model.NewResUserQuit(roomId, user.Uid)
	resMsg.To = uids
	e.SendQ <- resMsg
	return nil
}

func (e *MrCore) HandleReqSu(msg *model.Message, user *model.User) error {
	password := msg.Msg.(string)
	if password == e.Conf.SuSecret {
		user.IsSuperUser = true
	}
	return nil
}

func (e *MrCore) HandleReqNodeInfo(msg *model.Message, user *model.User) error {
	if !user.IsSuperUser {
		return fmt.Errorf("not super user")
	}
	var res model.CmdResNodeInfo
	res.RoomCnt = len(e.Room)
	res.UserCnt = len(e.User2Rooms)

	e.SendQ <- model.NewResNodeInfo([]int64{user.Uid}, &res)
	return nil
}

func (e *MrCore) HandleReqClear(msg *model.Message, user *model.User) error {
	if !user.IsSuperUser {
		return fmt.Errorf("not super user")
	}

	now := time.Now()
	shouldDel := make(map[int64]bool)
	for roomId, room := range e.Room {
		if roomId == model.DefaultRoomId {
			continue
		}
		shouldDrop := true
		if !room.Owner.IsDown || room.Owner.LastDownTime.Add(e.Conf.ClearInterval).Before(now) {
			shouldDrop = false
		} else {
			for _, v := range room.Users {
				if !v.IsDown || v.LastDownTime.Add(v.RestKickDuration).Before(now) {
					shouldDrop = false
					break
				}
			}
		}
		if shouldDrop {
			shouldDel[roomId] = true
		}
	}
	for k := range shouldDel {
		r := e.Room[k]
		var users []int64
		for uid := range r.Users {
			delete(e.User2Rooms[uid], r.ID)
			users = append(users, uid)
		}
		delete(e.Room, k)

		resMsg := model.NewResRoomDeleted(k)
		resMsg.To = users
		e.SendQ <- resMsg
	}
	return nil
}
