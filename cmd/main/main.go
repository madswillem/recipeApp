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
	server := server.NewServer(&server.Config{})
	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}

}
