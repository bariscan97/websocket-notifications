package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub
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
	username := c.Param("slug")

	page := c.Query("page")

	_, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	results, err := h.hub.Cache.GetNotifications(username, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.hub.Cache.ResetUnreadCount(username)

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

	userSlug := c.Param("slug")

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		Username: userSlug,
	}

	h.hub.Register <- cl

	go cl.writeMessage()
	go cl.readMessage(h.hub)
}
