package discord

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

const (
	DiscordAppID = "1393701411233202387"
)

var (
	joinButton = &client.Button{
		Label: "Приєднатись",
		Url:   "https://tblock.wtf",
	}
)

func Login() error {
	return client.Login(DiscordAppID)
}

func SetIdleActivity() error {
	now := time.Now()
	return client.SetActivity(client.Activity{
		State: "В головному меню",
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{joinButton},
	})
}

func SetPlayingActivity() error {
	now := time.Now()
	return client.SetActivity(client.Activity{
		State: "У грі",
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{joinButton},
	})
}
