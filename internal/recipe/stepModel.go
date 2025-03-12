package recipe

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type StepsStruct struct {
	ID           string    `db:"id" json:"id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	Step         string    `db:"step" json:"step"`
	RecipeID     string    `db:"recipe_id" json:"recipe_id"`
	TechniqueID  *string   `db:"technique_id" json:"technique_id,omitempty"`
	IngredientID *string   `db:"ingredient_id" json:"ingredient_id,omitempty"`
}

func (s *StepsStruct) Create(tx *sqlx.Tx) *error_handler.APIError {
	query := `INSERT INTO step (recipe_id, technique_id, ingredient_id, step)
			VAlUES (:recipe_id, :technique_id, :ingredient_id, :step)`

	_, db_err := tx.NamedExec(query, &s)
	if db_err != nil {
		tx.Rollback()
		return error_handler.New("Error creating steps: "+db_err.Error(), http.StatusInternalServerError, db_err)
	}

	return nil
}
