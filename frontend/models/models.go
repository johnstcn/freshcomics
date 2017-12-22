package models

import (
	"time"
)

type Comic struct {
	ID			int64
	Name      	string
	Title     	string
	SeenAt	 	time.Time
	NSFW 		bool
}

type ClickLog struct {
	ID			int64
	UpdateID	int64
	ClickedAt	time.Time
	Country		string
	Region		string
	City		string
	Zip			string
}
