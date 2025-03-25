package recipe

import (
	"errors"
	"time"
)

type IngredientsSchema struct {
	ID               string           `db:"id" json:"id"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	RecipeID         string           `db:"recipe_id" json:"recipe_id"`
	IngredientID     string           `db:"ingredient_id" json:"ingredient_id"`
	Amount           int              `db:"amount" json:"amount"`
	Unit             string           `db:"unit" json:"unit"`
	Name             string           `db:"name" json:"name"`
	NutritionalValue NutritionalValue `db:"nv" json:"nv"`
	Rating           RatingStruct     `db:"rating" json:"rating"`
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