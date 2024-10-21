package main


type QueueMessage struct {
	From    string `json:"from"`
	Content string `json:"content"`
}

type Message struct {
	User_slug string
	UnreadCount  string
}

type Emit struct {
	User_slug string
}