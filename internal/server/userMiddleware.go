package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/madswillem/recipeApp/internal/models"
)

func (s *Server) UserMiddleware(c *gin.Context) {
	var user models.UserModel
	cookie, ok := c.Cookie("user")
	if ok != nil || cookie == "" {
		err := user.Create(s.NewDB, c.ClientIP())
		if err != nil {
			error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		}
		c.SetCookie("user", user.Cookie, 31536000, "/", "localhost", false, true)
	} else {
		user.Cookie = cookie
		err := user.GetByCookie(s.NewDB)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		}
	}
	c.Set("user", user)
	c.Next()
}
