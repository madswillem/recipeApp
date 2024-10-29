package models

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
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

func (recipe *RecipeSchema) GetRecipeByID(db *sqlx.DB) *error_handler.APIError {
	err := db.Get(recipe, `SELECT recipes.*,
								rt.id AS "rating.id", rt.created_at AS "rating.created_at",
								rt.recipe_id AS "rating.recipe_id", rt.overall AS "rating.overall", rt.mon AS "rating.mon",
								rt.tue AS "rating.tue", rt.wed AS "rating.wed", rt.thu AS "rating.thu", rt.fri AS "rating.fri",
								rt.sat AS "rating.sat", rt.sun AS "rating.sun", rt.win AS "rating.win",
								rt.spr AS "rating.spr", rt.sum AS "rating.sum", rt.aut AS "rating.aut",
								rt.thirtydegree AS "rating.thirtydegree", rt.twentiedegree AS "rating.twentiedegree",
								rt.tendegree AS "rating.tendegree", rt.zerodegree AS "rating.zerodegree",
								rt.subzerodegree AS "rating.subzerodegree"
							FROM recipes
							LEFT JOIN rating rt ON rt.recipe_id = recipes.id
							WHERE recipes.id = $1`, recipe.ID)
	if err != nil {
		return error_handler.New("An error ocurred fetching the recipe: "+err.Error(), http.StatusInternalServerError, err)
	}

	err = db.Select(&recipe.Steps, `SELECT * FROM step WHERE recipe_id = $1`, recipe.ID)
	if err != nil {
		return error_handler.New("An error ocurred fetching the steps: "+err.Error(), http.StatusInternalServerError, err)
	}

	err = db.Select(&recipe.Ingredients, `SELECT recipe_ingredient.*, ingredient.name AS name
										FROM recipe_ingredient
										INNER JOIN ingredient ON ingredient.id = recipe_ingredient.ingredient_id
										WHERE recipe_id = $1`, recipe.ID)
	if err != nil {
		return error_handler.New("An error ocurred fetching the ingredients: "+err.Error(), http.StatusInternalServerError, err)
	}

	err = db.Select(&recipe.Diet, `
		SELECT diet.*
		FROM rel_diet_recipe rel
		JOIN diet  ON rel.diet_id = diet.id
		WHERE rel.recipe_id = $1
	`, recipe.ID)
	if err != nil {
		return error_handler.New("Error while getting nutritional values", http.StatusBadRequest, err)
	}

	return nil
}

func (recipe *RecipeSchema) UpdateSelected(change int, user *UserModel, db *sqlx.DB) *error_handler.APIError {
	apiErr := recipe.GetRecipeByID(db)
	if apiErr != nil {
		return apiErr
	}

	var attributes []string
	apiErr, attributes = UpdateRating(change)
	if apiErr != nil {
		return apiErr
	}

	fmt.Println(recipe.Rating.Overall)

	tx := db.MustBegin()
	tx.MustExec(`UPDATE "recipes" SET selects=selects + 1, version=version + 1 WHERE id=$1`, recipe.ID)

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
		return error_handler.New("Error updating rating", http.StatusInternalServerError, err)
	}

	if user == nil {
		return nil
	}
	apiErr = user.AddToGroup(db, recipe)

	return apiErr
}

func (recipe *RecipeSchema) CheckForRequiredFields() *error_handler.APIError {
	if recipe.Name == "" {
		return error_handler.New("missing recipe name", http.StatusBadRequest, errors.New("missing recipe name"))
	}
	if recipe.Ingredients == nil {
		return error_handler.New("missing recipe ingredients", http.StatusBadRequest, errors.New("missing recipe ingredients"))
	}
	if recipe.Steps == nil {
		return error_handler.New("missing recipe steps", http.StatusBadRequest, errors.New("missing recipe steps"))
	}
	for _, ingredient := range recipe.Ingredients {
		err := ingredient.CheckForRequiredFields()
		if err != nil {
			return error_handler.New(fmt.Sprintf("missing required field in ingredient %s %s", ingredient.Name, err.Error()), http.StatusBadRequest, err)
		}
	}

	return nil
}

func (recipe *RecipeSchema) Create(db *sqlx.DB) *error_handler.APIError {
	apiErr := recipe.CheckForRequiredFields()
	if apiErr != nil {
		return apiErr
	}

	for i := 0; i < len(recipe.Ingredients); i++ {
		recipe.Ingredients[i].Rating.DefaultRatingStruct(nil, &recipe.Ingredients[i].ID)
	}

	tx := db.MustBegin()
	// Insert recipe
	query := `INSERT INTO recipes (author, name, cuisine, yield, yield_unit, prep_time, cooking_time, selected, version)
              VALUES (:author, :name, :cuisine, :yield, :yield_unit, :prep_time, :cooking_time, :selected, :version) RETURNING id`
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return error_handler.New("Query error: "+err.Error(), http.StatusInternalServerError, err)
	}
	err = stmt.Get(&recipe.ID, recipe)
	stmt.Close()
	if err != nil {
		tx.Rollback()
		return error_handler.New("Dtabase error: "+err.Error(), http.StatusInternalServerError, err)
	}

	// Insert Rating
	recipe.Rating.DefaultRatingStruct(&recipe.ID, nil)
	query = `INSERT INTO rating (
				recipe_id, overall, mon, tue, wed, thu, fri, sat, sun, win, spr, sum, aut,
				thirtydegree, twentiedegree, tendegree, zerodegree, subzerodegree)
			VALUES (
				:recipe_id, :overall, :mon, :tue, :wed, :thu, :fri, :sat, :sun, :win, :spr, :sum, :aut,
				:thirtydegree, :twentiedegree, :tendegree, :zerodegree, :subzerodegree)`

	_, err = tx.NamedExec(query, recipe.Rating)
	if err != nil {
		tx.Rollback()
		return error_handler.New("Error inserting recipe: "+err.Error(), http.StatusInternalServerError, err)
	}

	// Insert Ingredient
	for _, ing := range recipe.Ingredients {
		ing.RecipeID = recipe.ID
		err := ing.Create(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	//Insert Steps
	for _, s := range recipe.Steps {
		s.RecipeID = recipe.ID
		err := s.Create(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	//Insert Diets
	for _, d := range recipe.Diet {
		var exists bool
		err = tx.Get(&exists, "SELECT EXISTS (SELECT 1 FROM diet WHERE id = $1);", d.ID)
		if err != nil {
			tx.Rollback()
			return error_handler.New("error while checking if diet exists", http.StatusInternalServerError, err)
		}
		if !exists {
			tx.Rollback()
			return error_handler.New("couldn't find diet "+d.ID, http.StatusNotFound, errors.New("couldn't find diet "+d.ID))
		}
		_, err = tx.Exec("INSERT INTO rel_diet_recipe (recipe_id, diet_id) VALUES ($1, $2)", recipe.ID, d.ID)
		if err != nil {
			tx.Rollback()
			return error_handler.New("error while inserting the relationship between diet and recipe", http.StatusInternalServerError, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return error_handler.New("Error creating recipe", http.StatusInternalServerError, err)
	}

	return nil
}
