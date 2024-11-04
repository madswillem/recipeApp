package initializers

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/models"
	"github.com/madswillem/recipeApp/internal/tools"
)

func InitDBonDev(db *sqlx.DB) error {
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
		err := recipe.Create(db)
		if err != nil {
			fmt.Printf("Recipe %s, Ingredient %s, err: %s", recipe.Name, "", err.Message)
		}
	}

	return nil
}
