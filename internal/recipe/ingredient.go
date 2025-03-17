package recipe

import (
	"net/http"

	"github.com/madswillem/recipeApp/internal/database"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type IngredientRepository struct {
}

func NewIngredientRepo() *IngredientRepository {
	return &IngredientRepository{}
}

func (ir *IngredientRepository) Create(ingredient *IngredientsSchema, db database.SQLDB) *error_handler.APIError {
	var err *error_handler.APIError
	ingredient.IngredientID, err = GetIngIDByName(ingredient.Name, db)
	if err != nil {
		return err
	}

	query := `INSERT INTO recipe_ingredient
    (recipe_id, ingredient_id, amount, unit)
    VALUES
    (:recipe_id, :ingredient_id, :amount, :unit)`

	_, db_err := db.NamedExec(query, &ingredient)
	if db_err != nil {
		return error_handler.New("Error creating "+ingredient.Name+": "+db_err.Error(), http.StatusInternalServerError, db_err)
	}

	return nil
}