package recipe

import "time"

type NutritionalValue struct {
	ID           string    `db:"id" json:"id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	IngredientID *string   `db:"ingredient_id" json:"ingredient_id,omitempty"`
	RecipeID     *string   `db:"recipe_id" json:"recipe_id,omitempty"`
	Kcal         float64   `db:"kcal" json:"kcal"`
	Kj           float64   `db:"kj" json:"kj"`
	Fat          float64   `db:"fat" json:"fat"`
	SaturatedFat float64   `db:"saturated_fat" json:"saturated_fat"`
	Carbohydrate float64   `db:"carbohydrate" json:"carbohydrate"`
	Sugar        float64   `db:"sugar" json:"sugar"`
	Protein      float64   `db:"protein" json:"protein"`
	Salt         float64   `db:"salt" json:"salt"`
	Nutriscore   string    `db:"nutriscore" json:"nutriscore"`
	Edited       bool      `json:"edited"`
}
