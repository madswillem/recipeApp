package models

import (
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type IngredientsSchema struct {
	ID           string    `db:"id" json:"id"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    RecipeID     string    `db:"recipe_id" json:"recipe_id"`
    IngredientID string    `db:"ingredient_id" json:"ingredient_id"`
    Amount       int64     `db:"amount" json:"amount"`
    Unit         string    `db:"unit" json:"unit"`
    Name         string    `db:"name" json:"name"`
	NutritionalValue NutritionalValue `db:"nv" json:"nv"`
	Rating	 RatingStruct `db:"rating" json:"rating"`
}

func (ingredient *IngredientsSchema) CheckForRequiredFields() error {
	if ingredient.Name == "" {
		return errors.New("missing name")
	}
	if ingredient.Amount == 0 {
		return errors.New("missing amount")
	}
	if ingredient.Unit == "" {
		return errors.New("missing measurement unit")
	}
	return nil
}

func (ingredient *IngredientsSchema) Create(tx *sqlx.Tx) *error_handler.APIError {
	var err *error_handler.APIError
	ingredient.IngredientID, err = GetIngIDByName(tx, ingredient.Name)
	if err != nil {
		return err
	}

	query := `INSERT INTO recipe_ingredient
    (recipe_id, ingredient_id, amount, unit)
    VALUES
    (:recipe_id, :ingredient_id, :amount, :unit)`

    _, db_err := tx.NamedExec(query, &ingredient)
    if db_err != nil {
		tx.Rollback()
        return error_handler.New("Error creating "+ingredient.Name+": "+db_err.Error(), http.StatusInternalServerError, db_err)
    }

	return nil
}
