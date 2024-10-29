package models

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type Filter struct {
	SearchText  *string   `db:"searchtext" json:"search"`
	NutriScore  *string   `db:"nutriscore" json:"nutriscore"`
	Name        *string   `db:"name" json:"name"`
	Cuisine     *string   `db:"cuisine" json:"cuisine"`
	PrepTime    *string   `db:"prep_time" json:"prep_time"`
	CookingTime *string   `db:"cooking_time" json:"cooking_time"`
	Ingredients *[]string `json:"ingredients"`
	Diets       *[]string `json:"diets"`
}

func (f *Filter) Filter(db *sqlx.DB) (*[]RecipeSchema, *error_handler.APIError) {
	recipes := []RecipeSchema{}
	var where []string
	var args []interface{}

	if f.SearchText != nil {
		where = append(where, `to_tsvector('english', recipes.name) @@ websearch_to_tsquery('english', $1)
					OR to_tsvector('english', ingredient.name) @@ websearch_to_tsquery('english', $1)
					OR to_tsvector('english', step.step) @@ websearch_to_tsquery('english', $1)`)
		args = append(args, f.SearchText)
	}
	if f.NutriScore != nil {
		args = append(args, f.NutriScore)
		where = append(where, fmt.Sprintf(`nutritional_value.nutriscore = :$%d`, len(args)))
	}
	if f.Cuisine != nil {
		args = append(args, f.Cuisine)
		where = append(where, fmt.Sprintf(`recipes.cuisine = $%d`, len(args)))
	}
	if f.PrepTime != nil {
		args = append(args, f.PrepTime)
		where = append(where, fmt.Sprintf(`recipes.prep_time <= $%d`, len(args)))
	}
	if f.CookingTime != nil {
		args = append(args, f.CookingTime)
		where = append(where, fmt.Sprintf(`recipes.cooking_time <= $%d`, len(args)))
	}
	if f.Ingredients != nil && len(*f.Ingredients) > 0 {
		for _, ing := range *f.Ingredients {
			args = append(args, ing)
			where = append(where, fmt.Sprintf(`ingredient.name = $%d`, len(args)))
		}
	}
	if f.Diets != nil {
		for _, d := range *f.Diets {
			args = append(args, d)
			where = append(where, fmt.Sprintf(`diet.id = $%d`, len(args)))
		}
	}

	query := `SELECT DISTINCT recipes.*
				FROM recipes
				LEFT JOIN recipe_ingredient ON recipes.id = recipe_ingredient.recipe_id
				LEFT JOIN ingredient ON ingredient.id = recipe_ingredient.ingredient_id
				LEFT JOIN nutritional_value ON recipes.id = nutritional_value.recipe_id
				LEFT JOIN step ON recipes.id = step.recipe_id
				JOIN rel_diet_recipe rel ON recipes.id = rel.recipe_id
				JOIN diet ON rel.diet_id = diet.id
				WHERE ` + strings.Join(where, " AND ")

	err := db.Select(&recipes, query, args...)
	if err != nil {
		return nil, error_handler.New("Dtabase error: "+err.Error(), http.StatusInternalServerError, err)
	}

	if len(recipes) <= 0 {
		return nil, nil
	}

	recipeMap := make(map[string]*RecipeSchema)
	for i := range recipes {
		recipeMap[recipes[i].ID] = &recipes[i]
	}
	id_array := make([]string, len(recipes))
	for i, r := range recipes {
		id_array[i] = r.ID
	}

	ingredients := []IngredientsSchema{}
	query, args, err = sqlx.In(`SELECT recipe_ingredient.*, ingredient.name FROM recipe_ingredient INNER JOIN ingredient ON ingredient.id = recipe_ingredient.ingredient_id WHERE recipe_ingredient.recipe_id IN (?)`, id_array)
	if err != nil {
		return nil, error_handler.New("error building ingredients query: "+err.Error(), http.StatusInternalServerError, err)
	}

	query = db.Rebind(query)

	err = db.Select(&ingredients, query, args...)
	if err != nil {
		return nil, error_handler.New("error fetching ingredients: "+query, http.StatusInternalServerError, err)
	}

	for _, ingredient := range ingredients {
		if recipe, found := recipeMap[ingredient.RecipeID]; found {
			recipe.Ingredients = append(recipe.Ingredients, ingredient)
		}
	}

	steps := []StepsStruct{}
	query, args, err = sqlx.In(`SELECT * FROM step WHERE step.recipe_id IN (?)`, id_array)
	if err != nil {
		return nil, error_handler.New("error fetching steps: "+err.Error(), http.StatusInternalServerError, err)
	}

	query = db.Rebind(query)

	err = db.Select(&steps, query, args...)
	if err != nil {
		return nil, error_handler.New("error fetching steps: "+err.Error(), http.StatusInternalServerError, err)
	}

	for _, step := range steps {
		if recipe, found := recipeMap[step.RecipeID]; found {
			recipe.Steps = append(recipe.Steps, step)
		}
	}

	return &recipes, nil
}
