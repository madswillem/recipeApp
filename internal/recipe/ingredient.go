package recipe

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type IngredientRepository interface {
	Create(ingredient *IngredientsSchema) *error_handler.APIError
}

type IngredientRepo struct {
	DB *sqlx.Tx
}

func NewIngredientRepo(db *sqlx.Tx) *IngredientRepo {
	return &IngredientRepo{
		DB: db,
	}
}

func (ir *IngredientRepo) Create(ingredient *IngredientsSchema) *error_handler.APIError {
	var err *error_handler.APIError
	ingredient.IngredientID, err = GetIngIDByName(ir.DB, ingredient.Name)
	if err != nil {
		return err
	}

	query := `INSERT INTO recipe_ingredient
    (recipe_id, ingredient_id, amount, unit)
    VALUES
    (:recipe_id, :ingredient_id, :amount, :unit)`

	_, db_err := ir.DB.NamedExec(query, &ingredient)
	if db_err != nil {
		ir.DB.Rollback()
		return error_handler.New("Error creating "+ingredient.Name+": "+db_err.Error(), http.StatusInternalServerError, db_err)
	}

	return nil
}