package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/madswillem/recipeApp/internal/recipe"
	"github.com/madswillem/recipeApp/internal/user"
)

func (s *Server) GetAll(c *gin.Context) {
	recipes, err := s.RecipeRepo.GetAllRecipes()
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}
	c.JSON(http.StatusOK, recipes)
}

func (s *Server) GetPopular(c *gin.Context) {
	recipes, err := s.RecipeRepo.GetByFilter(&recipe.Filter{})
	if err != nil {
		print(err.Errors)
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}
	c.JSON(http.StatusOK, recipes)
}

func (s *Server) AddRecipe(c *gin.Context) {
	var body recipe.RecipeSchema

	binderr := c.ShouldBindJSON(&body)
	if binderr != nil {
		error_handler.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{binderr})
		return
	}

	user := user.UserModel{}
	err := user.GetFromGinContext(c.Get("user"))
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	body.Build(user.ID)

	err = s.RecipeRepo.Create(&body)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	c.JSON(http.StatusCreated, body)
}

func (s *Server) AddIngredient(c *gin.Context) {
	var body recipe.IngredientDB

	binderr := c.ShouldBindJSON(&body)
	if binderr != nil {
		error_handler.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{binderr})
		return
	}

	err := body.Create(s.NewDB)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	c.JSON(http.StatusAccepted, body)
}

func (s *Server) UpdateRecipe(c *gin.Context) {
	user := user.UserModel{}
	usrerr := user.GetFromGinContext(c.Get("user"))
	if usrerr != nil {
		error_handler.HandleError(c, usrerr.Code, usrerr.Message, usrerr.Errors)
		return
	}

	ok, accesserr := s.Auth.AccessControl(user.ID, c.Param("id"), "update", s.RecipeRepo)
	if accesserr != nil {
		error_handler.HandleError(c, accesserr.Code, accesserr.Message, accesserr.Errors)
		return
	}
	if !ok {
		error_handler.HandleError(c, http.StatusUnauthorized, "User is not the owner of the recipe", []error{errors.New("user is not the owner of the recipe")})
		return
	}

	var body recipe.RecipeSchema
	err := c.ShouldBindJSON(&body)
	if err != nil {
		error_handler.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{err})
		return
	}

	log.Default().Println(body.Ingredients[0].Amount)

	updateerr := s.RecipeRepo.UpdateRecipe(c.Param("id"), &body)
	if updateerr != nil {
		error_handler.HandleError(c, updateerr.Code, updateerr.Message, updateerr.Errors)
		return
	}
}

func (s *Server) DeleteRecipe(c *gin.Context) {
	user := user.UserModel{}
	usrerr := user.GetFromGinContext(c.Get("user"))
	if usrerr != nil {
		error_handler.HandleError(c, usrerr.Code, usrerr.Message, usrerr.Errors)
		return
	}

	i := c.Param("id")
	owner, ownererr := s.RecipeRepo.GetRecipeAuthorbyID(i)
	if ownererr != nil {
		error_handler.HandleError(c, ownererr.Code, ownererr.Message, ownererr.Errors)
		return
	}
	if owner != user.ID {
		error_handler.HandleError(c, http.StatusUnauthorized, "User is not the owner of the recipe", []error{errors.New("user is not the owner of the recipe")})
		return
	}

	err := s.RecipeRepo.DeleteRecipe(i)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) GetById(c *gin.Context) {
	i := c.Param("id")

	result, err := s.RecipeRepo.GetRecipeByID(i)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
	}

	err = s.RecipeRepo.UpdateRecipeView(i)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
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
		error_handler.HandleError(c, http.StatusBadRequest, "Failed to read body", []error{err})
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
		error_handler.HandleError(c, apiErr.Code, apiErr.Message, apiErr.Errors)
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
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	err = s.RecipeRepo.UpdateRecipeSelect(c.Param("id"))
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	selectedErr := response.UpdateSelected(1, s.NewDB)
	if selectedErr != nil {
		error_handler.HandleError(c, selectedErr.Code, selectedErr.Message, selectedErr.Errors)
		return
	}

	err = user.AddToGroup(s.NewDB, response)
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) Deselect(c *gin.Context) {
	response, err := s.RecipeRepo.GetRecipeByID(c.Param("id"))
	if err != nil {
		error_handler.HandleError(c, err.Code, err.Message, err.Errors)
		return
	}

	selectedErr := response.UpdateSelected(-1, s.NewDB)
	if selectedErr != nil {
		error_handler.HandleError(c, selectedErr.Code, selectedErr.Message, selectedErr.Errors)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) Colormode(c *gin.Context) {
	switch c.Param("type") {
	case "get":
		cookie, err := c.Cookie("type")
		if err != nil {
			error_handler.HandleError(c, http.StatusBadRequest, "Cookie error", []error{err})
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
