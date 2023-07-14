package app

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Run(listener net.Listener) error {
	server := gin.Default()
	server.GET("/health_check", healthCheck)
	server.POST("/subscriptions", subscribe)

	return server.RunListener(listener)
}

func healthCheck(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

type FormData struct {
	Email string
	Name  string
}

func subscribe(ctx *gin.Context) {
	data := FormData{}
	email := ctx.PostForm("email")
	if email == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	data.Email = email

	name := ctx.PostForm("name")
	if name == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	data.Name = name
	ctx.Status(http.StatusOK)
}
