package models

import "time"

type Comic struct {
	ID			int64
	Name      	string
	Title     	string
	Published 	time.Time
	NSFW 		bool
}

