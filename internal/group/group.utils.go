package group

import (
	proto "github.com/isd-sgcu/rpkm67-go-proto/rpkm67/backend/group/v1"
	"github.com/isd-sgcu/rpkm67-model/model"
)

func ModelToProto(group *model.Group) *proto.Group {
	membersRPC := make([]*proto.UserInfo, len(group.Members))
	for i, m := range group.Members {
		membersRPC[i] = &proto.UserInfo{
			Id:        m.ID.String(),
			Firstname: m.Firstname,
			Lastname:  m.Lastname,
			ImageUrl:  m.PhotoUrl,
		}
	}

	return &proto.Group{
		Id:          group.ID.String(),
		LeaderID:    group.LeaderID.String(),
		Token:       group.Token,
		IsConfirmed: group.IsConfirmed,
		Members:     membersRPC,
	}
}

func UserToUserInfo(user *model.User) *proto.UserInfo {
	return &proto.UserInfo{
		Id:        user.ID.String(),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		ImageUrl:  user.PhotoUrl,
	}
}
