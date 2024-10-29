package main

import (
	"fmt"

	"github.com/madswillem/recipeApp/internal/initializers"
	"github.com/madswillem/recipeApp/internal/server"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {

	config := server.Config{
		//Innit:  []server.InnitFuncs{initializers.InitDBonDev},
	}
	server := server.NewServer(&config)
	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
