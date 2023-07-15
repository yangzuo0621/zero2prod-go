package app

import (
	"database/sql"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func Run(listener net.Listener, db *sql.DB) error {
	server := gin.Default()

	controller := Controller{DB: db}
	server.GET("/health_check", controller.healthCheck)
	server.POST("/subscriptions", controller.subscribe)

	return server.RunListener(listener)
}

type Controller struct {
	DB *sql.DB
}

func (controller *Controller) healthCheck(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

type FormData struct {
	Email string
	Name  string
}

func (controller *Controller) subscribe(ctx *gin.Context) {
	data := FormData{}
	email := ctx.PostForm("email")
	if email == "" {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	data.Email = email

	name := ctx.PostForm("name")
	if name == "" {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	data.Name = name

	id, err := uuid.NewV4()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = controller.DB.Exec(
		`
		INSERT INTO subscriptions (id, email, name, subscribed_at)
		VALUES ($1, $2, $3, $4)
		`,
		id,
		email,
		name,
		time.Now(),
	)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusOK)
}
