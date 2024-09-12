package ws

import (
	"net/http"
	"notifications/cacherepo"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub   *Hub
	cache cacherepo.IRedisClient
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) GetNotifications(c *gin.Context) {
	username := c.Param("username")

	page := c.Query("page")

	_, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	results, err := h.cache.GetNotifications(username, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"notifications": results,
	})
}

func (h *Handler) JoinWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Param("username")

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		Username: username,
	}

	m := &Message{
		UnreadCount: h.cache.GetUnreadCount(username),
		Username:    username,
	}

	h.hub.Register <- cl
	h.hub.Broadcast <- m

	go cl.writeMessage()
	go cl.readMessage(h.hub)
}
