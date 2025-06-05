package model

import "github.com/cloudwego/eino/schema"

type Request struct {
	History    []*schema.Message `json:"history"`
	UnreadData string            `json:"unread_data"`
}

type Response struct {
	Text string `json:"text"`
}
