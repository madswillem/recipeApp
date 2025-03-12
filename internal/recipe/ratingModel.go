package recipe

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/madswillem/recipeApp/internal/tools"
)

type RatingStruct struct {
	ID            string    `db:"id" json:"id"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	IngredientID  *string   `db:"ingredient_id" json:"ingredient_id,omitempty"`
	RecipeID      *string   `db:"recipe_id" json:"recipe_id,omitempty"`
	Overall       float64   `db:"overall" json:"overall"`
	Mon           float64   `db:"mon" json:"mon"`
	Tue           float64   `db:"tue" json:"tue"`
	Wed           float64   `db:"wed" json:"wed"`
	Thu           float64   `db:"thu" json:"thu"`
	Fri           float64   `db:"fri" json:"fri"`
	Sat           float64   `db:"sat" json:"sat"`
	Sun           float64   `db:"sun" json:"sun"`
	Win           float64   `db:"win" json:"win"`
	Spr           float64   `db:"spr" json:"spr"`
	Sum           float64   `db:"sum" json:"sum"`
	Aut           float64   `db:"aut" json:"aut"`
	ThirtyDegree  float64   `db:"thirtydegree" json:"thirtydegree"`
	TwentyDegree  float64   `db:"twentiedegree" json:"twentiedegree"`
	TenDegree     float64   `db:"tendegree" json:"tendegree"`
	ZeroDegree    float64   `db:"zerodegree" json:"zerodegree"`
	SubZeroDegree float64   `db:"subzerodegree" json:"subzerodegree"`
}

func (rating *RatingStruct) DefaultRatingStruct(recipe_id *string, ingredient_id *string) {
	rating.RecipeID = recipe_id
	rating.IngredientID = ingredient_id

	rating.Overall = 1000.0
	rating.Mon = 1000.0
	rating.Tue = 1000.0
	rating.Wed = 1000.0
	rating.Thu = 1000.0
	rating.Fri = 1000.0
	rating.Sat = 1000.0
	rating.Sun = 1000.0
	rating.Win = 1000.0
	rating.Spr = 1000.0
	rating.Sum = 1000.0
	rating.Aut = 1000.0
	rating.ThirtyDegree = 1000.0
	rating.TwentyDegree = 1000.0
	rating.TenDegree = 1000.0
	rating.ZeroDegree = 1000.0
	rating.SubZeroDegree = 1000.0
}

func UpdateRating(change int) (*error_handler.APIError, []string) {
	data, err := tools.GetCurrentData()

	if err != nil {
		return error_handler.New("fatal internal error", http.StatusInternalServerError, err), nil
	}

	factor := 1.1
	var result []string

	switch data.Day {
	case "Mon":
		result = append(result, fmt.Sprintf("mon = mon * %f", factor*float64(change)))
	case "Tue":
		result = append(result, fmt.Sprintf("tue = tue * %f", factor*float64(change)))
	case "Wed":
		result = append(result, fmt.Sprintf("wed = wed * %f", factor*float64(change)))
	case "Thu":
		result = append(result, fmt.Sprintf("thu = thu * %f", factor*float64(change)))
	case "Fri":
		result = append(result, fmt.Sprintf("fri = fri * %f", factor*float64(change)))
	case "Sat":
		result = append(result, fmt.Sprintf("sat = sat * %f", factor*float64(change)))
	case "Sun":
		result = append(result, fmt.Sprintf("sun = sun * %f", factor*float64(change)))
	default:
		return error_handler.New("fatal internal error", http.StatusInternalServerError, errors.New("fatal internal error")), nil
	}

	switch data.Season {
	case "Win":
		result = append(result, fmt.Sprintf("win = win * %f", factor*float64(change)))
	case "Spr":
		result = append(result, fmt.Sprintf("spr = spr * %f", factor*float64(change)))
	case "Sum":
		result = append(result, fmt.Sprintf("sum = sum * %f", factor*float64(change)))
	case "Aut":
		result = append(result, fmt.Sprintf("aut = aut * %f", factor*float64(change)))
	default:
		return error_handler.New("fatal internal error", http.StatusInternalServerError, errors.New("fatal internal error")), nil
	}

	switch data.Temp {
	case "subzerodegree":
		result = append(result, fmt.Sprintf("subzerodegree = subzerodegree * %f", factor*float64(change)))
	case "zerodegree":
		result = append(result, fmt.Sprintf("zerodegree = zerodegree * %f", factor*float64(change)))
	case "tendegree":
		result = append(result, fmt.Sprintf("tendegree = tendegree * %f", factor*float64(change)))
	case "twentiedegree":
		result = append(result, fmt.Sprintf("twentiedegree = twentiedegree * %f", factor*float64(change)))
	case "thirtydegree":
		result = append(result, fmt.Sprintf("thirtydegree = thirtydegree * %f", factor*float64(change)))
	}

	return nil, result
}
