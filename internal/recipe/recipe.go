package recipe

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type RecipeRepository interface {
	GetAllRecipes() ([]RecipeSchema, *error_handler.APIError)
	GetByFilter(f *Filter) ([]RecipeSchema, *error_handler.APIError)
	GetRecipeByID(id string) (*RecipeSchema, *error_handler.APIError)
	GetRecipeAuthorbyID(id string) (string, *error_handler.APIError)
	Create(recipe *RecipeSchema) *error_handler.APIError
	DeleteRecipe(id string) *error_handler.APIError
	UpdateRecipe(id string, recipe *RecipeSchema) *error_handler.APIError
	UpdateRecipeView(id string) *error_handler.APIError
	UpdateRecipeSelect(id string) *error_handler.APIError
	AddIngredient(id string, ingredient *IngredientsSchema) *error_handler.APIError
	DeleteIngredient(id string, ingredientID string) *error_handler.APIError
}

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

type RecipeRepo struct {
	DB *sqlx.DB
	IngRep *IngredientRepository
	StepRepo *StepRepository
	DietRepo *DietRepository
}

func NewRecipeRepo(db *sqlx.DB) *RecipeRepo {
	return &RecipeRepo{
		DB: db,
		IngRep: NewIngredientRepo(),
		StepRepo: NewStepRepo(),
		DietRepo: NewDietRepo(),
	}
}

func (rp *RecipeRepo) GetAllRecipes() ([]RecipeSchema, *error_handler.APIError) {
	recipes := []RecipeSchema{}

	err := rp.DB.Select(&recipes, `SELECT recipes.*,
								rt.id AS "rating.id", rt.created_at AS "rating.created_at",
								rt.recipe_id AS "rating.recipe_id", rt.overall AS "rating.overall", rt.mon AS "rating.mon",
								rt.tue AS "rating.tue", rt.wed AS "rating.wed", rt.thu AS "rating.thu", rt.fri AS "rating.fri",
								rt.sat AS "rating.sat", rt.sun AS "rating.sun", rt.win AS "rating.win",
								rt.spr AS "rating.spr", rt.sum AS "rating.sum", rt.aut AS "rating.aut",
								rt.thirtydegree AS "rating.thirtydegree", rt.twentiedegree AS "rating.twentiedegree",
								rt.tendegree AS "rating.tendegree", rt.zerodegree AS "rating.zerodegree",
								rt.subzerodegree AS "rating.subzerodegree"
								
							FROM recipes
							LEFT JOIN rating rt ON rt.recipe_id = recipes.id`)
	if err != nil {
		print(err.Error())
		return nil, error_handler.New("Error while getting recipes", http.StatusBadRequest, err)
	}

	if len(recipes) <= 0 {
		return nil, nil
	}

	apierr := rp.completeRecipes(recipes)
	if apierr != nil {
		return nil, apierr
	}

	return recipes, nil
}

func (rp *RecipeRepo) GetByFilter(f *Filter) ([]RecipeSchema, *error_handler.APIError) {
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

	query := fmt.Sprintf(
		`SELECT id, created_at, author, name, cuisine, yield, yield_unit, prep_time, cooking_time, version, selects, views
	    FROM (
	        SELECT DISTINCT ON (recipes.id) recipes.*, log.view_change
	        FROM recipes
	        LEFT JOIN recipe_ingredient ON recipes.id = recipe_ingredient.recipe_id
	        LEFT JOIN ingredient ON ingredient.id = recipe_ingredient.ingredient_id
	        LEFT JOIN nutritional_value ON recipes.id = nutritional_value.recipe_id
	        LEFT JOIN step ON recipes.id = step.recipe_id
	        LEFT JOIN recipe_selects_views_log log ON recipes.id = log.recipe_id
	        LEFT JOIN rel_diet_recipe rel ON recipes.id = rel.recipe_id
	        LEFT JOIN diet ON rel.diet_id = diet.id
	        %s
	        ORDER BY recipes.id, log.day DESC, log.view_change DESC
	    ) subquery
	    ORDER BY subquery.view_change DESC;`,
		func() string {
			if len(where) > 0 {
				return "WHERE " + strings.Join(where, " AND ")
			}
			return ""
		}(),
	)

	err := rp.DB.Select(&recipes, query, args...)
	if err != nil {
		return nil, error_handler.New("Dtabase error: "+err.Error(), http.StatusInternalServerError, err)
	}

	if len(recipes) <= 0 {
		return nil, nil
	}

	apierr := rp.completeRecipes(recipes)
	if apierr != nil {
		return nil, apierr
	}

	return recipes, nil
} 

func (rp *RecipeRepo) completeRecipes(recipes []RecipeSchema) *error_handler.APIError {
	// Prepare
	recipeMap := make(map[string]*RecipeSchema, len(recipes))
	for i := range recipes {
		recipeMap[recipes[i].ID] = &recipes[i]
	}
	id_array := make([]string, len(recipes))
	for i, r := range recipes {
		id_array[i] = r.ID
	}

	// Get ingredients
	ingredients := []IngredientsSchema{}
	query, args, err := sqlx.In(`SELECT recipe_ingredient.*, ingredient.name FROM recipe_ingredient INNER JOIN ingredient ON ingredient.id = recipe_ingredient.ingredient_id WHERE recipe_ingredient.recipe_id IN (?)`, id_array)
	if err != nil {
		return error_handler.New("error building ingredients query: "+err.Error(), http.StatusInternalServerError, err)
	}

	query = rp.DB.Rebind(query)

	err = rp.DB.Select(&ingredients, query, args...)
	if err != nil {
		return error_handler.New("error fetching ingredients: "+query, http.StatusInternalServerError, err)
	}

	for _, ingredient := range ingredients {
		if recipe, found := recipeMap[ingredient.RecipeID]; found {
			recipe.Ingredients = append(recipe.Ingredients, ingredient)
		}
	}

	// Get steps
	steps := []StepsStruct{}
	query, args, err = sqlx.In(`SELECT * FROM step WHERE step.recipe_id IN (?)`, id_array)
	if err != nil {
		return error_handler.New("error fetching steps: "+err.Error(), http.StatusInternalServerError, err)
	}

	query = rp.DB.Rebind(query)

	err = rp.DB.Select(&steps, query, args...)
	if err != nil {
		return error_handler.New("error fetching steps: "+err.Error(), http.StatusInternalServerError, err)
	}

	for _, step := range steps {
		if recipe, found := recipeMap[step.RecipeID]; found {
			recipe.Steps = append(recipe.Steps, step)
		}
	}



	//Get Diets
	var diets []struct {
		RecipeID      string            `db:"recipe_id"`
		ID            string            `db:"id" json:"id"`
		CreatedAt     time.Time         `db:"created_at" json:"created_at"`
		Name          string            `db:"name" json:"name"`
		Description   string            `db:"description" json:"description"`
		ExIngCategory []Category 		`json:"exingcategory"`
	}
	query, args, err = sqlx.In(`
		SELECT diet.*, rel.recipe_id
		FROM rel_diet_recipe rel
		JOIN diet ON rel.diet_id = diet.id
		WHERE rel.recipe_id IN (?)
	`, id_array)
	if err != nil {
		return error_handler.New("Error while getting diets", http.StatusInternalServerError, err)
	}

	query = rp.DB.Rebind(query)

	err = rp.DB.Select(&diets, query, args...)
	if err != nil {
		return error_handler.New("error fetching steps: "+err.Error(), http.StatusInternalServerError, err)
	}


	for _, rd := range diets {
		if recipe, exists := recipeMap[rd.RecipeID]; exists {
			recipe.Diet = append(recipe.Diet, DietSchema{
				ID:          rd.ID,
				CreatedAt:   rd.CreatedAt,
				Name:        rd.Name,
				Description: rd.Description,
			})
		}
	}

	return nil
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
	recipe.Rating.RecipeID = &recipe.ID
	fmt.Println(recipe.Rating)
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
		err := rp.IngRep.Create(&ing, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	//Insert Steps
	for _, s := range recipe.Steps {
		s.RecipeID = recipe.ID
		err := rp.StepRepo.Create(&s, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	//Insert Diets
	for _, d := range recipe.Diet {
		err := rp.DietRepo.Create(&d, recipe.ID, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return error_handler.New("Error creating recipe", http.StatusInternalServerError, err)
	}

	return nil
}

func (rp *RecipeRepo) DeleteRecipe(id string) *error_handler.APIError {
	result, err := rp.DB.Exec(`DELETE FROM public.recipes WHERE id = $1`, id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "22P02" { // 22P02 is the code for invalid_text_representation
			return error_handler.New(fmt.Sprintf(`Value "%s" is not an ID`, id), http.StatusBadRequest, err)
		}
		return error_handler.New(err.Error(), http.StatusInternalServerError, err)
	}

	if rows, _ := result.RowsAffected(); rows <= 0 {
		return error_handler.New("Recipe doesn't exist", http.StatusNotFound, nil)
	}

	return nil
}

func (rp *RecipeRepo) UpdateRecipe(id string, recipe *RecipeSchema) *error_handler.APIError {
	var setParts []string
	var args []interface{}

	if recipe.Name != "" {
		setParts = append(setParts, "name = $"+strconv.Itoa(len(args)+1))
		args = append(args, recipe.Name)
	}
	if recipe.Cuisine != "" {
		setParts = append(setParts, "cuisine = $"+strconv.Itoa(len(args)+1))
		args = append(args, recipe.Cuisine)
	}
	if recipe.Yield != 0 {
		setParts = append(setParts, "yield = $"+strconv.Itoa(len(args)+1))
		args = append(args, recipe.Yield)
	}
	if recipe.YieldUnit != "" {
		setParts = append(setParts, "yield_unit = $"+strconv.Itoa(len(args)+1))
		args = append(args, recipe.YieldUnit)
	}
	if recipe.PrepTime != "" {
		setParts = append(setParts, "prep_time = $"+strconv.Itoa(len(args)+1))
		args = append(args, recipe.PrepTime)
	}
	if recipe.CookingTime != "" {
		setParts = append(setParts, "cooking_time = $"+strconv.Itoa(len(args)+1))
		args = append(args, recipe.CookingTime)
	}

	if len(setParts) == 0 && len(recipe.Ingredients) == 0 {
		// No fields to body
		return error_handler.New("Nothing to update", http.StatusExpectationFailed, errors.New("nothing to update"))
	}

	tx := rp.DB.MustBegin()

	if len(setParts) > 0 {
		query := "UPDATE recipes SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(len(args)+1)
		args = append(args, id)

		_, err := tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return error_handler.New("Error Updating recipe", http.StatusInternalServerError, err)
		}
	}

	if len(recipe.Ingredients) != 0 {
		for _, ing := range recipe.Ingredients {
			err := rp.IngRep.Update(ing.ID, id, &ing, tx)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	err := tx.Commit()
	if err != nil {
		return error_handler.New("Error updating recipe", http.StatusInternalServerError, err)
	}

	return nil
}

func (rp *RecipeRepo) UpdateRecipeView(id string) *error_handler.APIError {
	_, err := rp.DB.Exec(`UPDATE recipes SET views = views + 1 WHERE id = $1`, id)
	if err != nil {
		return error_handler.New("Error updating views", http.StatusInternalServerError, err)
	}
	return nil
}

func (rp *RecipeRepo) UpdateRecipeSelect(id string) *error_handler.APIError {
	_, err := rp.DB.Exec(`UPDATE recipes SET selects = selects + 1 WHERE id = $1`, id)
	if err != nil {
		return error_handler.New("Error updating selects", http.StatusInternalServerError, err)
	}
	return nil
}

func (rp *RecipeRepo) AddIngredient(id string, ingredient *IngredientsSchema) *error_handler.APIError {
	ingredient.RecipeID = id
	return rp.IngRep.Create(ingredient, rp.DB)
}

func (rp *RecipeRepo) DeleteIngredient(id string, ingredientID string) *error_handler.APIError {
	return rp.IngRep.Delete(id, ingredientID, rp.DB)
}