package router

import (
	"fmt"
	"notifications/ws"
	"time"
	"os"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter(wsHandler *ws.Handler) {
	r = gin.Default()
	port := os.Getenv("PORT")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{fmt.Sprintf("http://localhost:%s", port)},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == fmt.Sprintf("http://localhost:%s", port)
		},
		MaxAge: 12 * time.Hour,
	}))

	r.GET("/ws/unreadcount/:username", wsHandler.GetNotifications)
	r.GET("/ws/:username", wsHandler.JoinWs)

}

func Start(addr string) error {
	return r.Run(addr)
}
