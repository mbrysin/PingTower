package model

import (
	"time"
)

type Site struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}
