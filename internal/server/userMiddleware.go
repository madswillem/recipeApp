package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/madswillem/recipeApp/internal/apierror"
)

func (s *Server) UserMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		tokenString, _ = c.Cookie("token")
		if tokenString == "" {
			apierror.HandleError(c, 401, "Authorization header required", []error{errors.New("authorization header required")})
			return
		}
	}

	user, err := s.Auth.Verify(s.NewDB, tokenString)
	if err != nil {
		err.Handle(c)
		return
	}
	c.Set("user", user)
	c.Next()
}
