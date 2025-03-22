package recipe

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func (ir *IngredientRepository) Update(id string, recipe_id string, ingredient *IngredientsSchema, db database.SQLDB) *error_handler.APIError {
	var setParts []string
	var args []interface{}

	if ingredient.IngredientID != "" {
		setParts = append(setParts, "ingredient_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, ingredient.IngredientID)
	}
	if ingredient.Amount != 0 {
		setParts = append(setParts, "amount = $"+strconv.Itoa(len(args)+1))
		args = append(args, ingredient.Amount)
	}
	if ingredient.Unit != "" {
		setParts = append(setParts, "unit = $"+strconv.Itoa(len(args)+1))
		args = append(args, ingredient.Unit)
	}

	if len(setParts) == 0 {
		// No fields to body
		return error_handler.New("Nothing to update", http.StatusExpectationFailed, errors.New("nothing to update"))
	}

	query := "UPDATE recipe_ingredient SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(len(args)+1) +" AND recipe_id = $" + strconv.Itoa(len(args)+2)
	args = append(args, id, recipe_id)

	log.Default().Println(query)

	_, err := db.Exec(query, args...)
	if err != nil {
		return error_handler.New("Error Updating recipe", http.StatusInternalServerError, err)
	}
	return nil
}

func (ir *IngredientRepository) Delete(id string, recipe_id string, db database.SQLDB) *error_handler.APIError {
	query := "DELETE FROM recipe_ingredient WHERE id = $1 AND recipe_id = $2"
	_, err := db.Exec(query, id, recipe_id)
	if err != nil {
		return error_handler.New("Error Deleting ingredient", http.StatusInternalServerError, err)
	}
	return nil
}