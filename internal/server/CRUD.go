package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madswillem/recipeApp/internal/apierror"
	"github.com/madswillem/recipeApp/internal/recipe"
	"github.com/madswillem/recipeApp/internal/user"
)

func (s *Server) GetAll(c *gin.Context) {
	recipes, err := s.RecipeRepo.GetAllRecipes()
	if err != nil {
		err.Handle(c)
		return
	}
	c.JSON(http.StatusOK, recipes)
}

func (s *Server) GetPopular(c *gin.Context) {
	recipes, err := s.RecipeRepo.GetByFilter(&recipe.Filter{})
	if err != nil {
		print(err.Errors)
		err.Handle(c)
		return
	}
	c.JSON(http.StatusOK, recipes)
}

func (s *Server) AddRecipe(c *gin.Context) {
	var body recipe.RecipeSchema

	binderr := c.ShouldBindJSON(&body)
	if binderr != nil {
		apierror.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{binderr})
		return
	}

	user := user.UserModel{}
	err := user.GetFromGinContext(c.Get("user"))
	if err != nil {
		err.Handle(c)
		return
	}

	body.Build(user.ID)

	err = s.RecipeRepo.Create(&body)
	if err != nil {
		err.Handle(c)
		return
	}

	c.JSON(http.StatusCreated, body)
}

func (s *Server) AddIngredient(c *gin.Context) {
	var body recipe.IngredientDB

	binderr := c.ShouldBindJSON(&body)
	if binderr != nil {
		apierror.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{binderr})
		return
	}

	err := body.Create(s.NewDB)
	if err != nil {
		err.Handle(c)
		return
	}

	c.JSON(http.StatusAccepted, body)
}

func (s *Server) UpdateRecipe(c *gin.Context) {
	user := user.UserModel{}
	usrerr := user.GetFromGinContext(c.Get("user"))
	if usrerr != nil {
		usrerr.Handle(c)
		return
	}

	ok, accesserr := s.Auth.AccessControl(user.ID, c.Param("id"), "update", s.RecipeRepo)
	if accesserr != nil {
		accesserr.Handle(c)
		return
	}
	if !ok {
		apierror.HandleError(c, http.StatusUnauthorized, "User is not the owner of the recipe", []error{errors.New("user is not the owner of the recipe")})
		return
	}

	var body recipe.RecipeSchema
	err := c.ShouldBindJSON(&body)
	if err != nil {
		apierror.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{err})
		return
	}

	log.Default().Println(body.Ingredients[0].Amount)

	updateerr := s.RecipeRepo.UpdateRecipe(c.Param("id"), &body)
	if updateerr != nil {
		updateerr.Handle(c)
		return
	}
}

func (s *Server) DeleteRecipe(c *gin.Context) {
	user := user.UserModel{}
	usrerr := user.GetFromGinContext(c.Get("user"))
	if usrerr != nil {
		usrerr.Handle(c)
		return
	}

	i := c.Param("id")
	owner, ownererr := s.RecipeRepo.GetRecipeAuthorbyID(i)
	if ownererr != nil {
		ownererr.Handle(c)
		return
	}
	if owner != user.ID {
		apierror.HandleError(c, http.StatusUnauthorized, "User is not the owner of the recipe", []error{errors.New("user is not the owner of the recipe")})
		return
	}

	err := s.RecipeRepo.DeleteRecipe(i)
	if err != nil {
		err.Handle(c)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) GetById(c *gin.Context) {
	i := c.Param("id")

	result, err := s.RecipeRepo.GetRecipeByID(i)
	if err != nil {
		err.Handle(c)
	}

	err = s.RecipeRepo.UpdateRecipeView(i)
	if err != nil {
		err.Handle(c)
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) Filter(c *gin.Context) {
	middleware_user, _ := c.Get("user")
	user, ok := middleware_user.(user.UserModel)
	if !ok {
		fmt.Println("type assertion failed")
	}

	var body recipe.Filter
	err := c.ShouldBindJSON(&body)
	if err != nil {
		apierror.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{err})
		return
	}

	if body.Diets == nil && ok {
		var diets []string
		err := s.NewDB.Select(&diets, `
			SELECT d.diet_id
			FROM rel_user_diet rel
			JOIN diet ON rel.diet_id = diet.diet_id
			WHERE rek.user_id = $1
		`, user.ID)
		if err != nil {
			print(err.Error())
		} else {
			body.Diets = &diets
		}
	}

	recipes, apiErr := s.RecipeRepo.GetByFilter(&body)
	if apiErr != nil {
		apiErr.Handle(c)
		return
	}

	c.JSON(http.StatusOK, recipes)
}

func (s *Server) Select(c *gin.Context) {
	middleware_user, _ := c.Get("user")
	user, ok := middleware_user.(user.UserModel)
	if !ok {
		fmt.Println("type assertion failed")
	}
	print(user.ID)

	response, err := s.RecipeRepo.GetRecipeByID(c.Param("id"))
	if err != nil {
		err.Handle(c)
		return
	}

	err = s.RecipeRepo.UpdateRecipeSelect(c.Param("id"))
	if err != nil {
		err.Handle(c)
		return
	}

	selectedErr := response.UpdateSelected(1, s.NewDB)
	if selectedErr != nil {
		selectedErr.Handle(c)
		return
	}

	err = user.AddToGroup(s.NewDB, response)
	if err != nil {
		err.Handle(c)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) Deselect(c *gin.Context) {
	response, err := s.RecipeRepo.GetRecipeByID(c.Param("id"))
	if err != nil {
		err.Handle(c)
		return
	}

	selectedErr := response.UpdateSelected(-1, s.NewDB)
	if selectedErr != nil {
		selectedErr.Handle(c)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) Colormode(c *gin.Context) {
	switch c.Param("type") {
	case "get":
		cookie, err := c.Cookie("type")
		if err != nil {
			apierror.HandleError(c, http.StatusBadRequest, "Cookie error", []error{err})
		}
		c.JSON(http.StatusOK, gin.H{"type": cookie})
	case "dark":
		c.SetCookie("type", "dark", 999999999999999999, "/", "localhost", false, true)
		c.Status(http.StatusAccepted)
	case "light":
		c.SetCookie("type", "light", 999999999999999999, "/", "localhost", false, true)
		c.Status(http.StatusAccepted)
	default:
		c.Status(http.StatusBadRequest)
	}
}
