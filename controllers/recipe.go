package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"rezeptapp.ml/goApp/initializers"
	"rezeptapp.ml/goApp/middleware"
	"rezeptapp.ml/goApp/models"
	"rezeptapp.ml/goApp/tools"
)

func GetAll(c *gin.Context) {
	var recipes []models.RecipeSchema

	result := initializers.DB.Preload("Ingredients").Find(&recipes)

	if result.Error != nil {
		panic(result.Error)
	}

	c.JSON(http.StatusOK, recipes)
}

func AddRecipe(c *gin.Context) {
	var body models.RecipeSchema
	id := tools.NewObjectId()

	err := c.Bind(&body)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Failed to read body",
			"errMessage": err.Error(),
		})
		return
	}

	body.ID = id
	for i:=0; i<len(body.Ingredients); i++ {   
		ingId := tools.NewObjectId()          // start of the execution block
        body.Ingredients[i].ID = ingId
    } 
	initializers.DB.Create(&body)
	result := initializers.DB.Save(&body)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Databaseerror",
			"errMessage": result.Error,
		})
		return
	}

	c.JSON(http.StatusCreated, body)
}

func GetById(c *gin.Context) {
	result := middleware.GetDataByID(c.Param("id"), c)

	c.JSON(http.StatusOK, result)
}

func Filter(c *gin.Context) {
	i, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        // ... handle error
        panic(err)
    }

	// find
	var result models.IngredientsSchema
	err = initializers.DB.Preload("Recipes").Find(&result, i).Error

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
	}

	c.JSON(http.StatusOK , result.Recipes)
}

func Select(c *gin.Context) {
	res := middleware.UpdateSelected(c.Param("id"), +1, c)

	c.JSON(http.StatusOK, res)
}

func Deselect(c *gin.Context) {
	res := middleware.UpdateSelected(c.Param("id"), -1, c)

	c.JSON(http.StatusOK, res)
}

func Colormode(c *gin.Context) {
	if c.Param("type") == "get" {
		cookie, err := c.Cookie("type")

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err": err,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"type": cookie,
		})
	} else if c.Param("type") == "dark" {
		c.SetCookie("type", "dark", 999999999999999999, "/", "localhost", false, true)
		c.Status(http.StatusAccepted)
	} else if c.Param("type") == "light" {
		c.SetCookie("type", "light", 999999999999999999, "/", "localhost", false, true)
		c.Status(http.StatusAccepted)
	} else {
		c.Status(http.StatusBadRequest)
	}
}
