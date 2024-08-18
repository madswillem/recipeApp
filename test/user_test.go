package test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp_Backend_Go/internal/database"
	"github.com/madswillem/recipeApp_Backend_Go/internal/models"
)

func parseDuration(durationStr string) (time.Duration, error) {
	var hour, min, sec int
	_, err := fmt.Sscanf(durationStr, "%d:%d:%d", &hour, &min, &sec)
	if err != nil {
		return 0, err
	}
	return time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second, nil
}

func TestSimilarity(t *testing.T) {
	t.Run("test similarity recipe_group recipe", func(t *testing.T) {
		db := database.ConnectToDB(&sqlx.Conn{})
		r := models.RecipeSchema{ID: "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd"}
		r.GetRecipeByID(db)
		rp := models.RecipeGroupSchema{}
		rp.Create(&r)

		fmt.Println(rp.Compare(&r))
	})
}
func TestCreate(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		r := models.RecipeSchema{
			ID:          "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
			CreatedAt:   time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
			Author:      "f85a98f8-2572-420a-9ae5-2c997ad96b6d",
			Name:        "Classic Spaghetti Carbonara",
			Cuisine:     "italian",
			Yield:       500,
			YieldUnit:   "g",
			PrepTime:    "01:00:00",
			CookingTime: "01:00:00",
			Selected:    0,
			Version:     0,
			Ingredients: []models.IngredientsSchema{
				{
					ID:           "69842c21-5832-4c64-9d27-2ffb8abd4617",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "8d7de19b-30f3-4cfd-ae93-c33a8f19a18d",
					Amount:       1,
					Unit:         "tsp",
					Name:         "salt",
				},
				{
					ID:           "185ae84d-4fe5-4328-ba1d-7af4434cb521",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "69332cc2-7b6f-42aa-be4d-c2ac2f2954c0",
					Amount:       400,
					Unit:         "g",
					Name:         "Spaghetti",
				},
				{
					ID:           "2de9c1c6-cc35-4038-8fbc-17029984f1d8",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "5e8cd4c6-51aa-42aa-ac24-ac3997c73341",
					Amount:       150,
					Unit:         "g",
					Name:         "Pancetta",
				},
				{
					ID:           "c2f50f80-71dd-4374-a856-bf417a26a5eb",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "ea3f9073-6a75-4625-80d1-19dc42aca7ef",
					Amount:       4,
					Unit:         "large",
					Name:         "Egg",
				},
				{
					ID:           "ed5fbfb6-2d2d-4467-82cf-9e7a97924724",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "db630404-6115-4ca1-91cd-f9ed8981676f",
					Amount:       100,
					Unit:         "g",
					Name:         "Parmesan cheese",
				},
				{
					ID:           "a4dd3925-e377-4380-8f0c-797d266b40e4",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "567e990a-20cf-4f85-974f-38189c0bb64b",
					Amount:       2,
					Unit:         "cloves",
					Name:         "Garlic",
				},
				{
					ID:           "07c807a0-15c2-4db5-8ca6-836499825c46",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "c1ac47d6-2126-48a4-ad73-75da637dee65",
					Amount:       1,
					Unit:         "tsp",
					Name:         "Black pepper",
				},
			},
			Steps: []models.StepsStruct{
				{
					ID:           "705897bb-6ec9-4d5f-adfc-0a7b4fa471dc",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Cook the spaghetti according to package directions until al dente. Reserve 1 cup of pasta water, then drain the pasta.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "13b29b7b-8ce8-44ba-90ae-c243c98da031",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "While the pasta cooks, heat a large skillet over medium heat and add the pancetta. Cook until crispy, then remove from heat and set aside.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "e8148c4f-6203-49e9-b50a-aa3a8545e808",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "In a bowl, whisk together the eggs and grated Parmesan cheese until well combined.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "926126e2-463b-436c-b574-5fbb646f82c8",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Return the skillet with pancetta to low heat. Add the minced garlic and cook until fragrant, about 1 minute.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "f508edbc-4c63-4f7e-949c-2f89422d7ad9",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Add the cooked pasta to the skillet and toss to combine with the pancetta and garlic.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "5e800272-2815-4220-ae2b-dc08c4ffc80b",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Remove the skillet from heat and quickly pour in the egg and cheese mixture, tossing rapidly to create a creamy sauce. If the sauce is too thick, add a little reserved pasta water until desired consistency is reached.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "539c001b-aa6d-4e20-9721-9a34eef5cccc",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Season with salt and freshly ground black pepper to taste. Serve immediately with extra Parmesan cheese on top, if desired.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
			},
		}
		rp := models.RecipeGroupSchema{}
		rp.Create(&r)
	})
}
func TestAdd(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		r := models.RecipeSchema{
			ID:          "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
			CreatedAt:   time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
			Author:      "f85a98f8-2572-420a-9ae5-2c997ad96b6d",
			Name:        "Classic Spaghetti Carbonara",
			Cuisine:     "italian",
			Yield:       500,
			YieldUnit:   "g",
			PrepTime:    "01:00:00",
			CookingTime: "01:00:00",
			Selected:    0,
			Version:     0,
			Ingredients: []models.IngredientsSchema{
				{
					ID:           "69842c21-5832-4c64-9d27-2ffb8abd4617",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "8d7de19b-30f3-4cfd-ae93-c33a8f19a18d",
					Amount:       1,
					Unit:         "tsp",
					Name:         "salt",
				},
				{
					ID:           "185ae84d-4fe5-4328-ba1d-7af4434cb521",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "69332cc2-7b6f-42aa-be4d-c2ac2f2954c0",
					Amount:       400,
					Unit:         "g",
					Name:         "Spaghetti",
				},
				{
					ID:           "2de9c1c6-cc35-4038-8fbc-17029984f1d8",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "5e8cd4c6-51aa-42aa-ac24-ac3997c73341",
					Amount:       150,
					Unit:         "g",
					Name:         "Pancetta",
				},
				{
					ID:           "c2f50f80-71dd-4374-a856-bf417a26a5eb",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "ea3f9073-6a75-4625-80d1-19dc42aca7ef",
					Amount:       4,
					Unit:         "large",
					Name:         "Egg",
				},
				{
					ID:           "ed5fbfb6-2d2d-4467-82cf-9e7a97924724",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "db630404-6115-4ca1-91cd-f9ed8981676f",
					Amount:       100,
					Unit:         "g",
					Name:         "Parmesan cheese",
				},
				{
					ID:           "a4dd3925-e377-4380-8f0c-797d266b40e4",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "567e990a-20cf-4f85-974f-38189c0bb64b",
					Amount:       2,
					Unit:         "cloves",
					Name:         "Garlic",
				},
				{
					ID:           "07c807a0-15c2-4db5-8ca6-836499825c46",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					IngredientID: "c1ac47d6-2126-48a4-ad73-75da637dee65",
					Amount:       1,
					Unit:         "tsp",
					Name:         "Black pepper",
				},
			},
			Steps: []models.StepsStruct{
				{
					ID:           "705897bb-6ec9-4d5f-adfc-0a7b4fa471dc",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Cook the spaghetti according to package directions until al dente. Reserve 1 cup of pasta water, then drain the pasta.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "13b29b7b-8ce8-44ba-90ae-c243c98da031",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "While the pasta cooks, heat a large skillet over medium heat and add the pancetta. Cook until crispy, then remove from heat and set aside.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "e8148c4f-6203-49e9-b50a-aa3a8545e808",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "In a bowl, whisk together the eggs and grated Parmesan cheese until well combined.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "926126e2-463b-436c-b574-5fbb646f82c8",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Return the skillet with pancetta to low heat. Add the minced garlic and cook until fragrant, about 1 minute.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "f508edbc-4c63-4f7e-949c-2f89422d7ad9",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Add the cooked pasta to the skillet and toss to combine with the pancetta and garlic.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "5e800272-2815-4220-ae2b-dc08c4ffc80b",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Remove the skillet from heat and quickly pour in the egg and cheese mixture, tossing rapidly to create a creamy sauce. If the sauce is too thick, add a little reserved pasta water until desired consistency is reached.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
				{
					ID:           "539c001b-aa6d-4e20-9721-9a34eef5cccc",
					CreatedAt:    time.Date(2024, 7, 24, 15, 49, 43, 879625000, time.UTC),
					Step:         "Season with salt and freshly ground black pepper to taste. Serve immediately with extra Parmesan cheese on top, if desired.",
					RecipeID:     "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd",
					TechniqueID:  nil,
					IngredientID: nil,
				},
			},
		}
		rp := models.RecipeGroupSchema{}
		rp.Create(&r)

		//fmt.Println(rp.PreperationDict)
		fmt.Println(rp.PrepTime)
		rp.Add(&models.RecipeSchema{
			Ingredients: []models.IngredientsSchema{
				{
					Name: "rice",
				},
				{
					Name: "tomato_puree",
				},
				{
					Name: "zucchine",
				},
			},
			Steps: []models.StepsStruct{
				{
					Step: "Put the rice into a pan and let fry til you have light krisp",
				},
				{
					Step: "Add and mix tomato puree until you have a nice red color",
				},
				{
					Step: "Add zucchini and serve",
				},
			},
			PrepTime:    "00:04:00",
			CookingTime: "00:04:00",
			Cuisine:     "Indien",
		})
		//fmt.Println(rp.PreperationDict)
		fmt.Println(rp.PrepTime)
	})
}
func TestMerge(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		rp := models.RecipeGroupSchema{
			Recipes: []models.RecipeSchema{
				{ID: "string1"},
			},
			IngredientDict: map[string]int{
				"hi":  1,
				"i":   2,
				"am":  3,
				"ben": 4,
			},
			IngredientVec: []float64{
				1, 1, 1, 1,
			},
			CuisineDict: map[string]int{
				"italien": 1,
			},
			CuisineVec: []float64{
				1,
			},
			PreperationDict: map[string]int{
				"hello": 1,
			},
			PreperationVec: []float64{
				1,
			},
			TechniquesDict: map[string]int{
				"0000hedfgiuha": 1,
			},
			TechniquesVec: []float64{
				1,
			},
			PrepTime:    time.Hour * 2,
			CookingTime: time.Hour * 3,
		}
		rp2 := models.RecipeGroupSchema{
			Recipes: []models.RecipeSchema{
				{ID: "string2"},
			},
			IngredientDict: map[string]int{
				"hi":      1,
				"you":     2,
				"are":     3,
				"timothe": 4,
			},
			IngredientVec: []float64{
				1, 1, 1, 1,
			},
			CuisineDict: map[string]int{
				"french": 1,
			},
			CuisineVec: []float64{
				1,
			},
			PreperationDict: map[string]int{
				"bye": 1,
			},
			PreperationVec: []float64{
				1,
			},
			TechniquesDict: map[string]int{
				"0000dhfapohedfgiuha": 1,
			},
			TechniquesVec: []float64{
				1,
			},
			PrepTime:    time.Minute * 4,
			CookingTime: time.Minute * 4,
		}

		expected := models.RecipeGroupSchema{
			Recipes: []models.RecipeSchema{
				{ID: "string1"},
				{ID: "string2"},
			},
			IngredientDict: map[string]int{
				"hi":      1,
				"i":       2,
				"am":      3,
				"ben":     4,
				"you":     5,
				"are":     6,
				"timothe": 7,
			},
			IngredientVec: []float64{
				1, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5,
			},
			CuisineDict: map[string]int{
				"italien": 1,
				"french":  2,
			},
			CuisineVec: []float64{
				0.5, 0.5,
			},
			PreperationDict: map[string]int{
				"hello": 1,
				"bye":   2,
			},
			PreperationVec: []float64{
				0.5, 0.5,
			},
			TechniquesDict: map[string]int{
				"0000hedfgiuha":       1,
				"0000dhfapohedfgiuha": 2,
			},
			TechniquesVec: []float64{
				0.5, 0.5,
			},
			PrepTime:    time.Minute * 62,
			CookingTime: time.Minute * 92,
		}

		rp.Merge(&rp2)

		if !reflect.DeepEqual(rp.Recipes, expected.Recipes) {
			t.Errorf("Expected %+v but got %+v", expected.Recipes, rp.Recipes)
		}
		if !reflect.DeepEqual(rp.IngredientDict, expected.IngredientDict) {
			t.Errorf("Expected %+v but got %+v", expected.IngredientDict, rp.IngredientDict)
		}
		if !reflect.DeepEqual(rp.IngredientVec, expected.IngredientVec) {
			t.Errorf("Expected %+v but got %+v", expected.IngredientVec, rp.IngredientVec)
		}
		if !reflect.DeepEqual(rp.CuisineDict, expected.CuisineDict) {
			t.Errorf("Expected %+v but got %+v", expected.CuisineDict, rp.CuisineDict)
		}
		if !reflect.DeepEqual(rp.CuisineVec, expected.CuisineVec) {
			t.Errorf("Expected %+v but got %+v", expected.CuisineVec, rp.CuisineVec)
		}
		if !reflect.DeepEqual(rp.PreperationDict, expected.PreperationDict) {
			t.Errorf("Expected %+v but got %+v", expected.PreperationDict, rp.PreperationDict)
		}
		if !reflect.DeepEqual(rp.PreperationVec, expected.PreperationVec) {
			t.Errorf("Expected %+v but got %+v", expected.PreperationVec, rp.PreperationVec)
		}
		if !reflect.DeepEqual(rp.TechniquesDict, expected.TechniquesDict) {
			t.Errorf("Expected %+v but got %+v", expected.TechniquesDict, rp.TechniquesDict)
		}
		if !reflect.DeepEqual(rp.TechniquesVec, expected.TechniquesVec) {
			t.Errorf("Expected %+v but got %+v", expected.TechniquesVec, rp.TechniquesVec)
		}
		if !reflect.DeepEqual(rp.PrepTime, expected.PrepTime) {
			t.Errorf("Expected %+v but got %+v", expected.PrepTime, rp.PrepTime)
		}
		if !reflect.DeepEqual(rp.CookingTime, expected.CookingTime) {
			t.Errorf("Expected %+v but got %+v", expected.CookingTime, rp.CookingTime)
		}
	})
}
func TestUser_Add(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		db := database.ConnectToDB(&sqlx.Conn{})
		r := models.RecipeSchema{ID: "aa85daf1-dbc5-462d-a6fe-3fbb358b08dd"}
		r.GetRecipeByID(db)
		user := models.UserModel{ID: "a5e4a4fb-1b70-4cfc-8350-55decee258ab"}
		err := user.AddToGroup(db, &r)
		if err != nil {
			t.Errorf("Error: %+v\n", err)
		}
	})
}

func TestGetUser(t *testing.T) {
	t.Run("by cookie", func(t *testing.T) {
		db := database.ConnectToDB(&sqlx.Conn{})
		u := models.UserModel{Cookie: "A2H0PFDRIuj5G8YC6qdq"}
		err := u.GetByCookie(db)
		if err != nil {
			t.Errorf("%+v", err)
			return
		}
	})
}
