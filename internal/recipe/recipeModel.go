package recipe

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/apierror"
)

type RecipeSchema struct {
	ID               string    `db:"id"`
	CreatedAt        time.Time `db:"created_at"`
	Author           string    `db:"author"`
	Name             string    `db:"name"`
	Cuisine          string    `db:"cuisine"`
	Yield            int       `db:"yield"`
	YieldUnit        string    `db:"yield_unit"`
	PrepTime         string    `db:"prep_time"`
	CookingTime      string    `db:"cooking_time"`
	Selects          int       `db:"selects"`
	Views            int       `db:"views"`
	Version          int64     `db:"version"`
	Ingredients      []IngredientsSchema
	Diet             []DietSchema
	NutritionalValue NutritionalValue
	Rating           RatingStruct `db:"rating"` 
	Steps            []StepsStruct
}

func (recipe *RecipeSchema) UpdateSelected(change int, db *sqlx.DB) *apierror.APIError {
	attributes, apiErr := UpdateRating(change)
	if apiErr != nil {
		return apiErr
	}

	fmt.Println(recipe.Rating.Overall)

	tx := db.MustBegin()

	query := `UPDATE rating SET ` + strings.Join(attributes, ",") + ` WHERE recipe_id = $1;`
	println(query)
	tx.MustExec(query, recipe.ID)

	query = `WITH updated_values AS (
        SELECT (
			mon + tue + wed + thu + fri + sat + sun +
			win + spr + sum + aut +
			thirtydegree + twentiedegree + tendegree + zerodegree + subzerodegree
		) / 16.0 AS average
        FROM rating
        WHERE recipe_id = $1
    )
    UPDATE rating
    SET overall = (SELECT average FROM updated_values)
    WHERE recipe_id = $1;`
	tx.MustExec(query, recipe.ID)

	err := tx.Commit()
	if err != nil {
		return apierror.New("Error updating rating", http.StatusInternalServerError, err)
	}
	

	return apiErr
}

func (recipe *RecipeSchema) checkForRequiredFields() *apierror.APIError {
	if recipe.Name == "" {
		return apierror.New("missing recipe name", http.StatusBadRequest, errors.New("missing recipe name"))
	}
	if recipe.Ingredients == nil {
		return apierror.New("missing recipe ingredients", http.StatusBadRequest, errors.New("missing recipe ingredients"))
	}
	if recipe.Steps == nil {
		return apierror.New("missing recipe steps", http.StatusBadRequest, errors.New("missing recipe steps"))
	}

	return nil
}

func (recipe *RecipeSchema) Build(authorid string) *apierror.APIError {
	// Ensure Recipe has all required fields
	apiErr := recipe.checkForRequiredFields()
	if apiErr != nil {
		return apiErr
	}
	recipe.Rating.DefaultRatingStruct(&recipe.ID, nil)

	// Ensure all ingredients have all required fields
	for _, ingredient := range recipe.Ingredients {
		err := ingredient.CheckForRequiredFields()
		if err != nil {
			return apierror.New(fmt.Sprintf("missing required field in ingredient %s %s", ingredient.Name, err.Error()), http.StatusBadRequest, err)
		}
	}
	for i := 0; i < len(recipe.Ingredients); i++ {
		recipe.Ingredients[i].Rating.DefaultRatingStruct(nil, &recipe.Ingredients[i].ID)
	}

	return nil
}