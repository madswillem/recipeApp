package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/madswillem/recipeApp/internal/recipe"
	"github.com/madswillem/recipeApp/internal/tools"
)

type UserModel struct {
	ID           string    `database:"id"`
	CreatedAt    time.Time `db:"created_at"`
	Email        string    `db:"email" json:"email"`
	Password     string    `db:"password" json:"-"`
	LastLogin    time.Time `database:"last_login"`
	Cookie       string    `database:"cookie"`
	IP           string    `database:"ip"`
	RecipeGroups []RecipeGroupSchema
	Groups       []byte       `database:"groups"`
	Settings     UserSettings `database:"settings"`
}
type UserSettings struct {
	Allergies []*string `database:"allergies"`
	Diet      recipe.DietSchema      `gorm:"polymorphic:Owner"`
}

func (user *UserModel) GetByCookie(db *sqlx.DB) *error_handler.APIError {
	err := db.Get(user, `SELECT id, created_at, ip, groups FROM "user" WHERE cookie = $1`, user.Cookie)
	if err != nil {
		return error_handler.New("database error", http.StatusInternalServerError, err)
	}
	err = json.Unmarshal(user.Groups, &user.RecipeGroups)
	if err != nil {
		return error_handler.New("Error unmarshaling groups", http.StatusInternalServerError, err)
	}
	return nil
}

func (user *UserModel) CheckIfExistsByCookie(db *sqlx.DB) bool {
	found := false
	err := db.Get(found, "SELECT EXISTS(SELECT * FROM user WHERE cookie = $1) AS found;", user.Cookie)
	if err != nil {
		error_handler.New("database error", http.StatusInternalServerError, err)
	}
	return found
}

func (user *UserModel) Create(db *sqlx.DB, ip string) *error_handler.APIError {
	user.LastLogin = time.Now()
	user.IP = ip
	for {
		user.Cookie = tools.RandomString(20)
		if !user.CheckIfExistsByCookie(db) {
			break
		}
	}

	query := `INSERT INTO "user" (cookie, ip) VALUES (:cookie, :ip) RETURNING id`
	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return error_handler.New("Query error: "+err.Error(), http.StatusInternalServerError, err)
	}
	err = stmt.Get(&user.ID, user)
	stmt.Close()
	if err != nil {
		return error_handler.New("Error inserting user: "+err.Error(), http.StatusInternalServerError, err)
	}

	return nil
}

func (user *UserModel) GetFromGinContext(data any, exists bool) *error_handler.APIError {
	if data == nil {
		return error_handler.New("User not logged in", http.StatusUnauthorized, errors.New("user not logged in"))
	}
	userData, ok := data.(UserModel)
	if !ok {
		return error_handler.New("User not of type UserModel", http.StatusUnauthorized, errors.New("user not of type UserModel"))
	}
	*user = userData
	return nil
}

func (user *UserModel) AddGroup(db *sqlx.DB, r *recipe.RecipeSchema) *error_handler.APIError {
	rp := RecipeGroupSchema{}
	rp.Create(r)
	user.RecipeGroups = append(user.RecipeGroups, rp)

	v, err := json.Marshal(user.RecipeGroups)
	if err != nil {
		return error_handler.New("Failed to marshal recipe groups", http.StatusInternalServerError, err)
	}

	db.MustExec(`UPDATE "user" SET groups = $1 WHERE id = $2`, v, user.ID)
	return nil
}
func (user *UserModel) AddToGroup(db *sqlx.DB, r *recipe.RecipeSchema) *error_handler.APIError {
	var groups []byte
	err := db.Get(&groups, `SELECT "groups" FROM "user" WHERE id=$1`, user.ID)
	if err != nil {
		return error_handler.New("Error fetching recipe_groups", http.StatusInternalServerError, err)
	}
	err = json.Unmarshal(groups, &user.RecipeGroups)
	if err != nil {
		return error_handler.New("Error unmarshaling groups", http.StatusInternalServerError, err)
	}
	if len(user.RecipeGroups) < 1 {
		return user.AddGroup(db, r)
	}

	group_ranking := make([]struct {
		Group *RecipeGroupSchema
		Sim   float64
	}, len(user.RecipeGroups))

	for i := range user.RecipeGroups {
		group_ranking[i].Group = &user.RecipeGroups[i]
		group_ranking[i].Sim = user.RecipeGroups[i].Compare(r)
	}

	group_addble := make([]struct {
		Group *RecipeGroupSchema
		Sim   float64
	}, 0)
	fmt.Printf("%+v\n", group_ranking)
	for i := range group_ranking {
		if group_ranking[i].Sim >= .9 {
			group_addble = append(group_addble, group_ranking[i])
		}
	}
	fmt.Printf("%+v\n", group_ranking)

	if len(group_addble) > 1 {
		for i := 1; i < len(group_addble); i++ {
			group_addble[0].Group.Merge(group_addble[i].Group)
		}
	}

	group_addble[0].Group.Add(r)
	user.Groups, err = json.Marshal(user.RecipeGroups)
	if err != nil {
		return error_handler.New("failed to marshal", http.StatusInternalServerError, err)
	}
	_, err = db.Exec(`UPDATE "user" SET groups = $1 WHERE id = $2`, user.Groups, user.ID)
	if err != nil {
		return error_handler.New("database error", http.StatusInternalServerError, err)
	}

	return nil
}

// Using db to extend an existing db like a search to show recipes similar to your intrests
func (user *UserModel) GetRecomendation(db *sqlx.DB) (*error_handler.APIError, []recipe.RecipeSchema) {
	return nil, nil
}
