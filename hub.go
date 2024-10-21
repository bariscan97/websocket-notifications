package main

type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Emitter    chan *Emit
	Cache      *RedisClient
}

func NewHub(cache *RedisClient) *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Emitter:    make(chan *Emit, 5),
		Cache:      cache,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.Register:
			h.Clients[cl.Username] = cl
			h.Emitter <- &Emit{
				User_slug: cl.Username,
			}
		case cl := <-h.Unregister:
			close(cl.Message)
			delete(h.Clients, cl.Username)
		case m := <-h.Emitter:
			cl, ok := h.Clients[m.User_slug]
			if ok {
				cl.Message <- &Message{
					User_slug:   m.User_slug,
					UnreadCount: h.Cache.GetUnreadCount(cl.Username),
				}
			}
		}
	}
}
