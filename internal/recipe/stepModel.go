package recipe

import (
	"time"
)

type StepsStruct struct {
	ID           string    `db:"id" json:"id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	Step         string    `db:"step" json:"step"`
	RecipeID     string    `db:"recipe_id" json:"recipe_id"`
	TechniqueID  *string   `db:"technique_id" json:"technique_id,omitempty"`
	IngredientID *string   `db:"ingredient_id" json:"ingredient_id,omitempty"`
}
