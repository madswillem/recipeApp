package models

import (
	"fmt"
	"net/http"

	"github.com/madswillem/recipeApp_Backend_Go/internal/error_handler"
	"github.com/madswillem/recipeApp_Backend_Go/internal/initializers"
)

type RecipeGroupSchema struct {
	ID                uint   `gorm:"primarykey"`
	UserID		  uint
	Recipes           []*RecipeSchema  `gorm:"many2many:recipe_recipegroups"` 
	AvrgIngredients   []Avrg `gorm:"foreignKey:GroupID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AvrgCuisine       []Avrg `gorm:"foreignKey:GroupID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AvrgVegetarien    float64
	AvrgVegan         float64
	AvrgLowCal        float64
	AvrgLowCarb       float64
	AvrgKeto          float64
	AvrgPaleo         float64
	AvrgLowFat        float64
	AvrgFoodCombining float64
	AvrgWholeFood     float64
}

type Avrg struct {
	ID              uint `gorm:"primarykey"`
	GroupID		uint
	Name            string
	Percentige      float64
}

func (group *RecipeGroupSchema) GetRecipeGroupByID(reqData map[string]bool) *error_handler.APIError {
	err := initializers.DB.Find(&group, "ID = ?", group.ID).Error
	if err != nil {
		return error_handler.New("database error", http.StatusInternalServerError, err)
	}
	return nil
}
func GetAllRecipeGroups() (*error_handler.APIError, []RecipeGroupSchema) {
	var groups []RecipeGroupSchema
	err := initializers.DB.Find(&groups).Error
	if err != nil {
		return error_handler.New("database error", http.StatusInternalServerError, err), nil
	}
	return nil, groups
}
func (groups *RecipeGroupSchema) Update() *error_handler.APIError {
	err := initializers.DB.Save(&groups).Error
	if err != nil {
		return error_handler.New("database error", http.StatusInternalServerError, err)
	}
	return nil
}
func (group *RecipeGroupSchema) AddRecipeToGroup(recipe *RecipeSchema) {
	for _, name := range recipe.Ingredients {
		added := false
		for _, avrgName := range group.AvrgIngredients {
			if name.Ingredient == avrgName.Name {
				avrgName.Percentige += 1
				added = true
			}
		}
		if !added {
			group.AvrgIngredients = append(group.AvrgIngredients, Avrg{Name: name.Ingredient, Percentige: 1})
		}
	}
	for _, cuisine := range group.AvrgCuisine {
		if recipe.Cuisine == cuisine.Name {
			cuisine.Percentige += 1
		}
	}
	switch {
	case recipe.Diet.Vegetarien:
		group.AvrgVegetarien += 1
	case recipe.Diet.Vegan:
		group.AvrgVegan += 1
	case recipe.Diet.LowCarb:
		group.AvrgLowCarb += 1
	case recipe.Diet.Keto:
		group.AvrgKeto += 1
	case recipe.Diet.Paleo:
		group.AvrgPaleo += 1
	case recipe.Diet.LowFat:
		group.AvrgLowFat += 1
	case recipe.Diet.FoodCombining:
		group.AvrgFoodCombining += 1
	case recipe.Diet.WholeFood:
		group.AvrgWholeFood += 1
	}
	group.Recipes = append(group.Recipes, recipe)
	group.Update()
}

func GroupNew(recipe *RecipeSchema) RecipeGroupSchema {
	new := RecipeGroupSchema{}
	new.Recipes = append(new.Recipes, recipe)
	for _, ing := range recipe.Ingredients {
		new.AvrgIngredients = append(new.AvrgIngredients, Avrg{Name: ing.Ingredient, Percentige: 1})
	}
	new.AvrgCuisine = append(new.AvrgCuisine, Avrg{Name: recipe.Cuisine, Percentige: 1})
	switch {
		case recipe.Diet.Vegetarien: new.AvrgVegetarien = 1
		case recipe.Diet.Vegan: new.AvrgVegan = 1
		case recipe.Diet.LowCal: new.AvrgLowCal = 1
		case recipe.Diet.LowCarb: new.AvrgLowCarb = 1
		case recipe.Diet.Keto: new.AvrgKeto = 1
		case recipe.Diet.Paleo: new.AvrgPaleo = 1
		case recipe.Diet.LowFat: new.AvrgLowFat = 1
		case recipe.Diet.FoodCombining: new.AvrgFoodCombining = 1
		case recipe.Diet.WholeFood: new.AvrgWholeFood = 1
	}
	fmt.Println("New Group: ")
	fmt.Println(recipe)

	return new
}
type SimiliarityGroupRecipe struct {
	Group RecipeGroupSchema
	Similarity float64
}
func SortSimilarity( groups []SimiliarityGroupRecipe ) []SimiliarityGroupRecipe {
	len := len(groups)
	for i := 0; i < len-1; i++ {
		for j := 0; j < len-i-1; j++ {
			if groups[j].Similarity > groups[j+1].Similarity {
				groups[j], groups[j+1] = groups[j+1], groups[j]
			}
		}
	}
	return groups
}