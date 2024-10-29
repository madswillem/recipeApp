package initializers

import (
	"encoding/json"
	"fmt"

	"github.com/madswillem/recipeApp/internal/models"
	"github.com/madswillem/recipeApp/internal/server"
	"github.com/madswillem/recipeApp/internal/tools"
)

func InitDBonDev(s *server.Server) error {
	var recipes []models.RecipeSchema
	expected_return_string, err := tools.ReadFileAsString("./test/testdata/create/100_recipes.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(expected_return_string), &recipes)
	if err != nil {
		return err
	}

	for _, recipe := range recipes {
		err := recipe.Create(s.NewDB)
		if err != nil {
			fmt.Printf("Recipe %s, Ingredient %s, err: %s", recipe.Name, "", err.Message)
		}
	}

	return nil
}
