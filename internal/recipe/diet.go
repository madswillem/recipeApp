package recipe

import (
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type DietRepository struct {
}

func NewDietRepo() *DietRepository {
	return &DietRepository{}
}

func (dr *DietRepository) Create(diet *DietSchema, recipeid string, db *sqlx.Tx) *error_handler.APIError {
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS (SELECT 1 FROM diet WHERE id = $1);", diet.ID)
	if err != nil {
		return error_handler.New("error while checking if diet exists", http.StatusInternalServerError, err)
	}
	if !exists {
		return error_handler.New("couldn't find diet "+diet.ID, http.StatusNotFound, errors.New("couldn't find diet "+diet.ID))
	}
	_, err = db.Exec("INSERT INTO rel_diet_recipe (recipe_id, diet_id) VALUES ($1, $2)", recipeid, diet.ID)
	if err != nil {
		return error_handler.New("error while inserting the relationship between diet and recipe", http.StatusInternalServerError, err)
	}
	return nil
}