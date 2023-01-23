package main

import (
	"github.com/gin-gonic/gin"
	"rezeptapp.ml/goApp/controllers"
	"rezeptapp.ml/goApp/initializers"
)

func init()  {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main()  {
	r := gin.Default()
	r.POST("/create", controllers.AddRecipe)
	r.GET("/get", controllers.GetAll)
	r.Run()
}