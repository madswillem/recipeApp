// filepath: /home/mads/Documents/recipeApp/internal/recipe/step.go
package recipe

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type StepRepository struct {
}

func NewStepRepo() *StepRepository {
	return &StepRepository{}
}

func (s *StepRepository) Create(step *StepsStruct,tx *sqlx.Tx) *error_handler.APIError {
	query := `INSERT INTO step (recipe_id, technique_id, ingredient_id, step)
			VAlUES (:recipe_id, :technique_id, :ingredient_id, :step)`

	_, db_err := tx.NamedExec(query, &step)
	if db_err != nil {
		tx.Rollback()
		return error_handler.New("Error creating steps: "+db_err.Error(), http.StatusInternalServerError, db_err)
	}

	return nil
}