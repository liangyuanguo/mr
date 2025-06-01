package utility

import (
	"liangyuanguo/aw/mr/pkg/model"
)

func AddMem(r *model.Room, u *model.User) {
	r.Users[u.Uid] = u
}

func RemMem(r *model.Room, u *model.User) {
	delete(r.Users, u.Uid)
}

func NewRoom(name string, owner *model.User) *model.Room {
	r := &model.Room{
		Name:     name,
		Owner:    owner,
		Users:    make(map[int64]*model.User),
		IsPublic: true,
	}

	AddMem(r, owner)
	return r
}
