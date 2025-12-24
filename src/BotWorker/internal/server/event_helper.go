package server

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/plugins"
)

func processEventIDs(event *onebot.Event) {
	if event.Platform != "qqguild" {
		return
	}

	database := plugins.GlobalDB
	if database == nil {
		return
	}

	event.EnsureIDs(
		func(openID string) (int64, error) {
			return db.GetUserIDByOpenID(database, openID)
		},
		func(openID string) (int64, error) {
			return db.GetGroupIDByOpenID(database, openID)
		},
		func() (int64, error) {
			return db.GetMaxUserIDPlusOne(database)
		},
		func() (int64, error) {
			return db.GetMaxGroupIDPlusOne(database)
		},
		func(userID, targetID int64, openID, nickname, avatar string) error {
			return db.CreateUserWithTargetID(database, userID, targetID, openID, nickname, avatar)
		},
		func(groupID, targetID int64, openID, name string) error {
			return db.CreateGroupWithTargetID(database, groupID, targetID, openID, name)
		},
	)
}
