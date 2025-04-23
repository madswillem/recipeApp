package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madswillem/recipeApp/internal/apierror"
	"github.com/madswillem/recipeApp/internal/user"
)

func (s *Server) GetRecommendation(c *gin.Context) {
	middleware_user, _ := c.Get("user")
	user, ok := middleware_user.(user.UserModel)
	if !ok {
		fmt.Println("type assertion failed")
	}

	err := user.GetByCookie(s.NewDB)
	if err != nil {
		err.Handle(c)
		return
	}

	err, recipes := user.GetRecomendation(s.NewDB)
	if err != nil {
		err.Handle(c)
		return
	}

	c.JSON(http.StatusOK, recipes)
}
func (s *Server) CreateGroup(c *gin.Context) {
	middleware_user, _ := c.Get("user")
	u, ok := middleware_user.(user.UserModel)
	if !ok {
		fmt.Println("type assertion failed")
	}

	r, apiErr := s.RecipeRepo.GetRecipeByID("aa85daf1-dbc5-462d-a6fe-3fbb358b08dd")
	if apiErr != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, apiErr)
		return
	}

	rp := user.RecipeGroupSchema{}
	rp.Create(r)
	u.RecipeGroups = append(u.RecipeGroups, rp)

	v, err := json.Marshal(u.RecipeGroups)
	if err != nil {
		apierror.HandleError(c, http.StatusInternalServerError, "Couldnt Marshal recipe group", []error{err})
	}

	s.NewDB.MustExec(`UPDATE "user" SET groups = $1 WHERE id = $2`, v, u.ID)
	c.JSON(http.StatusAccepted, u)
}
