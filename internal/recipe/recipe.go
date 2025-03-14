package recipe

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/pkg/errors"
)

type RecipeRepository interface {
	//GetAllRecipes() ([]recipe.RecipeSchema, *error_handler.APIError)
	GetRecipeByID(id string) (*RecipeSchema, *error_handler.APIError)
	GetRecipeAuthorbyID(id string) (string, *error_handler.APIError)
	Create(recipe *RecipeSchema) *error_handler.APIError
}

type RecipeRepo struct {
	DB *sqlx.DB
}

func NewRecipeRepo(db *sqlx.DB) *RecipeRepo {
	return &RecipeRepo{
		DB: db,
	}
}

func (rp *RecipeRepo) GetRecipeByID(id string) (*RecipeSchema, *error_handler.APIError) {
	recipe := &RecipeSchema{}
	err := rp.DB.Get(recipe, `SELECT recipes.*,
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
							WHERE recipes.id = $1`, id)
	if err != nil {
		return nil, error_handler.New("An error ocurred fetching the recipe: "+err.Error(), http.StatusInternalServerError, err)
	}

	err = rp.DB.Select(&recipe.Steps, `SELECT * FROM step WHERE recipe_id = $1`, id)
	if err != nil {
		return nil, error_handler.New("An error ocurred fetching the steps: "+err.Error(), http.StatusInternalServerError, err)
	}

	err = rp.DB.Select(&recipe.Ingredients, `SELECT recipe_ingredient.*, ingredient.name AS name
										FROM recipe_ingredient
										INNER JOIN ingredient ON ingredient.id = recipe_ingredient.ingredient_id
										WHERE recipe_id = $1`, recipe.ID)
	if err != nil {
		return nil, error_handler.New("An error ocurred fetching the ingredients: "+err.Error(), http.StatusInternalServerError, err)
	}

	err = rp.DB.Select(&recipe.Diet, `
		SELECT diet.*
		FROM rel_diet_recipe rel
		JOIN diet  ON rel.diet_id = diet.id
		WHERE rel.recipe_id = $1
	`, recipe.ID)
	if err != nil {
		return nil, error_handler.New("Error while getting nutritional values", http.StatusBadRequest, err)
	}

	return recipe, nil
}

func (rp *RecipeRepo) GetRecipeAuthorbyID(id string) (string, *error_handler.APIError) {
	var owner string
	err := rp.DB.Get(&owner, `SELECT author FROM recipes WHERE id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", error_handler.New("Recipe doesn't exist", http.StatusNotFound, err)
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "22P02" { // 22P02 is the code for invalid_text_representation
			log.Printf("Potential SQL injection or invalid input detected, with input %s", id)
			return "", error_handler.New(fmt.Sprintf(`Value "%s" is not an ID`, id), http.StatusBadRequest, err)
		}
		return "", error_handler.New("Error while getting author", http.StatusInternalServerError, err)
	}
	return owner, nil
}

func (rp *RecipeRepo) Create(recipe *RecipeSchema) *error_handler.APIError {
	tx := rp.DB.MustBegin()
	// Insert recipe
	query := `INSERT INTO recipes (author, name, cuisine, yield, yield_unit, prep_time, cooking_time, version)
              VALUES (:author, :name, :cuisine, :yield, :yield_unit, :prep_time, :cooking_time, :version) RETURNING id`
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