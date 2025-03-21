package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

func (s *Server) UserMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		tokenString, _ = c.Cookie("token")
		if tokenString == "" {
			error_handler.HandleError(c, 401, "Authorization header required", []error{errors.New("authorization header required")})
			return
		}
	}

	err, user := s.Auth.Verify(s.NewDB, tokenString)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}
	c.Set("user", user)
	c.Next()
}
