package recipe

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/database"
	"github.com/madswillem/recipeApp/internal/error_handler"
)

type IngredientDB struct {
	ID               string           `db:"id" json:"id"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	Name             string           `db:"name" json:"name"`
	StandardUnit     string           `db:"standard_unit" json:"standard_unit,omitempty"`
	NdbNumber        int64            `db:"ndb_number" json:"ndb_number,omitempty"`
	Category         string           `db:"category" json:"category,omitempty"`
	FdicID           int64            `db:"fdic_id" json:"fdic_id,omitempty"`
	NutritionalValue NutritionalValue `json:"nv"`
	Rating           RatingStruct     `json:"rating"`
}
type Category struct {
	ID   string `db:"id"`
	Name string `db:"name" json:"name"`
}

func (ingredient *IngredientDB) Create(db *sqlx.DB) *error_handler.APIError {
	tx := db.MustBegin()
	// Create ingredient
	query := `INSERT INTO ingredient (name, standard_unit, ndb_number, category, fdic_id)
              VALUES (:name, :standard_unit, :ndb_number, :category, :fdic_id) RETURNING id`
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return error_handler.New("Query error: "+err.Error(), http.StatusInternalServerError, err)
	}
	err = stmt.Get(&ingredient.ID, ingredient)
	stmt.Close()
	if err != nil {
		tx.Rollback()
		return error_handler.New("Dtabase error: "+err.Error(), http.StatusInternalServerError, err)
	}

	// Create Rating
	ingredient.Rating.DefaultRatingStruct(nil, &ingredient.ID)
	query = `INSERT INTO rating (
				ingredient_id, overall, mon, tue, wed, thu, fri, sat, sun, win, spr, sum, aut,
				thirtydegree, twentiedegree, tendegree, zerodegree, subzerodegree)
			VALUES (
				:ingredient_id, :overall, :mon, :tue, :wed, :thu, :fri, :sat, :sun, :win, :spr, :sum, :aut,
				:thirtydegree, :twentiedegree, :tendegree, :zerodegree, :subzerodegree)`

	_, err = tx.NamedExec(query, ingredient.Rating)
	if err != nil {
		tx.Rollback()
		return error_handler.New("Error inserting recipe: "+err.Error(), http.StatusInternalServerError, err)
	}

	tx.Commit()
	return nil
}

func GetIngIDByName(name string, db database.SQLDB) (string, *error_handler.APIError) {
	var id string
	err := db.QueryRow("SELECT id FROM ingredient WHERE LOWER(name) = LOWER($1)", name).Scan(&id)
	if err != nil {
		return "", error_handler.New("database error getting "+name+" : "+err.Error(), http.StatusInternalServerError, err)
	}

	return id, nil
}
