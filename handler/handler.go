package handler

import (
	"net/http"
	"strconv"

	"github.com/cs3238-tsuzu/coding_challenge_03/model"
	"github.com/gin-gonic/gin"
)

// Handler is a struct for handler
type Handler struct {
	UserController model.UserController
	handler        http.Handler
}

// NewHandler initializes a handler for Hello world
func NewHandler(db model.DB) *Handler {
	router := gin.Default()

	handler := &Handler{
		handler: router,
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!!",
		})
	})

	router.GET("/users", func(c *gin.Context) {
		l, err := handler.UserController.ListUsers()

		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")

			return
		}

		c.JSON(http.StatusOK, l)
	})

	router.GET("/users/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.String(http.StatusBadRequest, "invalid id")

			return
		}
		res, err := handler.UserController.GetUser(id)

		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")

			return
		}

		c.JSON(http.StatusOK, res)
	})

	router.POST("/users", func(c *gin.Context) {
		type parameterType struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		var param parameterType

		if err := c.BindJSON(&param); err != nil {
			c.String(http.StatusBadRequest, "bad request")

			return
		}

		u, err := handler.UserController.NewUser(param.Name, param.Email)

		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")

			return
		}

		c.JSON(http.StatusCreated, u)
	})

	router.PUT("/users/:id", func(c *gin.Context) {
		var user model.User

		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.String(http.StatusBadRequest, "invalid id")

			return
		}

		if err := c.BindJSON(&user); err != nil {
			c.String(http.StatusBadRequest, "bad request")

			return
		}
		user.ID = id

		res, err := handler.UserController.UpdateUser(&user)

		if err != nil {
			if err == model.ErrNoUser {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "not found",
				})

				return
			}

			c.String(http.StatusInternalServerError, "internal server error")

			return
		}

		c.JSON(http.StatusOK, res)
	})

	router.DELETE("/users/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.String(http.StatusBadRequest, "invalid id")

			return
		}

		err = handler.UserController.DeleteUser(id)

		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")

			return
		}

		c.Status(http.StatusNoContent)
	})

	return handler
}

// GetHandler returns http.Handler
func (h *Handler) GetHandler() http.Handler {
	return h.handler
}
