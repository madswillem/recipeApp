package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/database"
	"github.com/madswillem/recipeApp/internal/recipe"
	"github.com/madswillem/recipeApp/internal/server"
	"github.com/madswillem/recipeApp/internal/tools"
	"github.com/madswillem/recipeApp/internal/user"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func assertRecipesEqual(t *testing.T, expected recipe.RecipeSchema, actual recipe.RecipeSchema) {
	if actual.ID == "" || expected.ID == "" {
		t.Errorf("%s (actual.ID) != %s (expected.ID)", actual.ID, expected.ID)

		return
	}

	// Check if the lengths of actual.Ingredients and expected.Ingredients are equal
	if len(actual.Ingredients) != len(expected.Ingredients) {
		t.Errorf("Expected %d ingredients but got %d", len(expected.Ingredients), len(actual.Ingredients))
		return
	}

	// Check if the lengths of actual.Steps and expected.Steps are equal
	if len(actual.Steps) != len(expected.Steps) {
		t.Errorf("Expected %d ingredients but got %d", len(expected.Steps), len(actual.Steps))
		return
	}

	var errors []string

	// Compare each ingredient in the actual recipe
	less := func(a, b recipe.IngredientsSchema) bool {
		return a.ID < b.ID // Sorting by Name (you can add more criteria if needed)
	}

	// Sort both lists
	sort.Slice(actual.Ingredients, func(i, j int) bool { return less(actual.Ingredients[i], actual.Ingredients[j]) })
	sort.Slice(expected.Ingredients, func(i, j int) bool { return less(actual.Ingredients[i], actual.Ingredients[j]) })

	// Compare the sorted lists
	diff := cmp.Diff(actual.Ingredients, actual.Ingredients, cmpopts.SortSlices(less))

	if diff != "" {
		t.Errorf("Ingredients are not the same: %s", diff)
	}

	// Compare Steps
	for num, step := range actual.Steps {
		expectedStep := expected.Steps[num]

		if step.Step != expectedStep.Step {
			errors = append(errors, fmt.Sprintf("Expected step %s but got %s", expectedStep.Step, step.Step))
		}
		if step.TechniqueID != expectedStep.TechniqueID {
			errors = append(errors, fmt.Sprintf("Expected technique %s but got %s", *expectedStep.TechniqueID, *step.TechniqueID))
		}
	}

	// Compare other recipe properties
	if actual.Name != expected.Name {
		errors = append(errors, fmt.Sprintf("Expected Name %s but got %s", expected.Name, actual.Name))
	}
	if actual.PrepTime != expected.PrepTime {
		errors = append(errors, fmt.Sprintf("Expected prep_time %s but got %s", expected.PrepTime, actual.PrepTime))
	}
	if actual.CookingTime != expected.CookingTime {
		errors = append(errors, fmt.Sprintf("Expected cooking_time %s but got %s", expected.CookingTime, actual.CookingTime))
	}
	if actual.NutritionalValue != expected.NutritionalValue {
		errors = append(errors, fmt.Sprintf("Expected NutritionalValue %v but got %v", expected.NutritionalValue, actual.NutritionalValue))
	}
	if actual.Rating.Overall != expected.Rating.Overall {
		errors = append(errors, fmt.Sprintf("Expected recipe rating %v but got %v", expected.Rating, actual.Rating))
	}

	if len(errors) > 0 {
		t.Error(strings.Join(errors, "\n"))
	}
}

func TestServer_AddRecipe(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("mads"),
		postgres.WithPassword("1234"),
		postgres.BasicWaitStrategies(),
		postgres.WithInitScripts("./testdata/innit-db.sql"),
		postgres.WithSQLDriver("pgx"),
	)
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
	if err != nil {
		t.Fatal(err)
	}

	dbURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	s := server.Server{NewDB: database.ConnectToDB(&sqlx.Conn{}, dbURL)}

	type testCase struct {
		name           string
		requestBody    string
		user 		 user.UserModel
		expectedStatus int
		expectedBody   string
	}

	testCases := []testCase{
		{
			name:           "add recipe with all required fields",
			requestBody:    "./testdata/create/add_recipe_with_all_required_fields/body.json",
			user:           user.UserModel{ID: "f85a98f8-2572-420a-9ae5-2c997ad96b6d"},
			expectedStatus: http.StatusCreated,
			expectedBody:   "./testdata/create/add_recipe_with_all_required_fields/expected_return.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)

			completeRequestFilePath, err := filepath.Abs(tc.requestBody)
			if err != nil {
				t.Fatal(err)
			}

			reqBody, err := tools.ReadFileAsString(completeRequestFilePath)
			if err != nil {
				t.Error(err)
				return
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user", tc.user)

			s.AddRecipe(c)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d but got %d. \n Body: %s", tc.expectedStatus, w.Code, w.Body.String())
			}

			if tc.expectedBody != "" {
				var response recipe.RecipeSchema
				err = json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Fatal(err)
				}

				completeExpectedFilePath, err := filepath.Abs(tc.expectedBody)
				if err != nil {
					t.Fatal(err)
				}

				var expectedReturn recipe.RecipeSchema
				expectedBody, err := tools.ReadFileAsString(completeExpectedFilePath)
				if err != nil {
					t.Fatal(err)
				}
				err = json.Unmarshal([]byte(expectedBody), &expectedReturn)
				if err != nil {
					t.Fatal(err)
				}

				assertRecipesEqual(t, expectedReturn, response)
			}
			t.Cleanup(func() {
				err = container.Restore(ctx)
				if err != nil {
					fmt.Printf("Error restoring container: %s\n", err.Error())
				}
			})
		})
	}
}

func TestServer_GetById(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("mads"),
		postgres.WithPassword("1234"),
		postgres.BasicWaitStrategies(),
		postgres.WithInitScripts("./testdata/innit-db.sql"),
		postgres.WithSQLDriver("pgx"),
	)
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
	if err != nil {
		t.Fatal(err)
	}

	dbURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	s := server.Server{NewDB: database.ConnectToDB(&sqlx.Conn{}, dbURL)}

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "get recipe by id: c4ef5707-1577-4f8c-99ef-0f492e82b895",
			id:             "c4ef5707-1577-4f8c-99ef-0f492e82b895",
			expectedStatus: http.StatusOK,
			expectedBody:   "./testdata/get/getbyid.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest(http.MethodGet, "/getbyid", nil)
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}

			s.GetById(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d but got %d. \n Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectedBody != "" {
				var response recipe.RecipeSchema
				err = json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Fatal(err)
				}

				completeExpectedFilePath, err := filepath.Abs(tt.expectedBody)
				if err != nil {
					t.Fatal(err)
				}

				var expectedReturn recipe.RecipeSchema
				expectedBody, err := tools.ReadFileAsString(completeExpectedFilePath)
				if err != nil {
					t.Fatal(err)
				}
				err = json.Unmarshal([]byte(expectedBody), &expectedReturn)
				if err != nil {
					t.Fatal(err)
				}

				assertRecipesEqual(t, expectedReturn, response)
			}
			t.Cleanup(func() {
				err = container.Restore(ctx)
				if err != nil {
					fmt.Printf("Error restoring container: %s\n", err.Error())
				}
			})
		})
	}
}

func TestServer_Delete(t *testing.T) {
	container, ctx := InitTestContainer(t)
	URL, err := container.ConnectionString(*ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	s := server.Server{NewDB: database.ConnectToDB(&sqlx.Conn{}, URL)}

	tests := []struct {
		name           string
		id             string
		user           user.UserModel
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "delete recipe with id c4ef5707-1577-4f8c-99ef-0f492e82b895",
			id:             "c4ef5707-1577-4f8c-99ef-0f492e82b895",
			user:           user.UserModel{ID: "f85a98f8-2572-420a-9ae5-2c997ad96b6d"},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "delete nonexisting recipe",
			id:             "c5ef5707-1577-4f8c-99ef-0f492e82b895",
			user:           user.UserModel{ID: "f85a98f8-2572-420a-9ae5-2c997ad96b6d"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"errMessage":"Recipe doesn't exist","errors":"sql: no rows in result set"}`,
		},
		{
			name:           "test sql injection",
			id:             `c5ef5707-1577-4f8c-99ef-0f492e82b895"; SELECT * FROM recipes;`,
			user:           user.UserModel{ID: "f85a98f8-2572-420a-9ae5-2c997ad96b6d"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"errMessage":"Value \"c5ef5707-1577-4f8c-99ef-0f492e82b895\"; SELECT * FROM recipes;\" is not an ID","errors":"pq: invalid input syntax for type uuid: \"c5ef5707-1577-4f8c-99ef-0f492e82b895\"; SELECT * FROM recipes;\""}`,
		},
		{
			name:           "delete recipe with invalid UUID format",
			id:             "invalid-uuid-format",
			user:           user.UserModel{ID: "f85a98f8-2572-420a-9ae5-2c997ad96b6d"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"errMessage":"Value \"invalid-uuid-format\" is not an ID","errors":"pq: invalid input syntax for type uuid: \"invalid-uuid-format\""}`,
		},
		{
			name:           "delete recipe with empty ID",
			id:             "",
			user:           user.UserModel{ID: "f85a98f8-2572-420a-9ae5-2c997ad96b6d"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"errMessage":"Value \"\" is not an ID","errors":"pq: invalid input syntax for type uuid: \"\""}`,
		},
		{
			name:           "delete recipe with unauthorized user",
			id:             "c4ef5707-1577-4f8c-99ef-0f492e82b895",
			user:           user.UserModel{ID: "wrong-user-id"},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"errMessage":"User is not the owner of the recipe","errors":"user is not the owner of the recipe"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest(http.MethodDelete, "/delete", nil)
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			c.Set("user", tt.user)

			s.DeleteRecipe(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d but got %d. \n Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectedBody != "" {
				if w.Body.String() != tt.expectedBody {
					t.Errorf("Expected Body: \n %s \n but got: \n %s", tt.expectedBody, w.Body.String())
				}
			} else if w.Body.String() != "" {
				t.Errorf("Unexpected response body: %s \n", w.Body.String())
			}

			t.Cleanup(func() {
				err = container.Restore(*ctx)
				if err != nil {
					fmt.Printf("Error restoring container: %s\n", err.Error())
				}
			})
		})
	}
}
