package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/madswillem/gocron"
	"github.com/madswillem/recipeApp/internal/auth"
	"github.com/madswillem/recipeApp/internal/database"
	"github.com/madswillem/recipeApp/internal/apierror"
	"github.com/madswillem/recipeApp/internal/initializers"
	"github.com/madswillem/recipeApp/internal/recipe"
	"github.com/madswillem/recipeApp/internal/user"
	"github.com/madswillem/recipeApp/internal/workers"
	views "github.com/madswillem/recipeApp/web/view"
)

const MethodGET = "GET"
const MethodPost = "POST"

type Auth interface {
	Login(c *gin.Context, db *sqlx.DB)
	Signup(c *gin.Context, db *sqlx.DB)
	Verify(db *sqlx.DB, tokenString string) (user.UserModel, *apierror.APIError)
	Logout(c *gin.Context)
	AccessControl(sub string, obj string, act string, repo recipe.RecipeRepository) (bool, *apierror.APIError)
}

type InnitFuncs func(*Server) error
type ExtraControllers struct {
	Function   func(*gin.Context)
	Middleware func(*gin.Context)
	Route      string
	Method     string
}

type Config struct {
	Innit       []InnitFuncs
	Auth        Auth
	Controllers []ExtraControllers
}

type Server struct {
	port       int
	NewDB      *sqlx.DB
	Registry   *gocron.Registry
	RecipeRepo recipe.RecipeRepository
	Auth       Auth
	config     *Config
}

func NewServer(config *Config) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	initializers.LoadEnvVariables()
	NewServer := &Server{
		port: port,
		NewDB: database.ConnectToDB(
			&sqlx.Conn{},
			fmt.Sprintf(
				"user=%s password=%s database=%s sslmode=disable",
				os.Getenv("POSTGRES_USER"),
				os.Getenv("POSTGRES_PASSWORD"),
				os.Getenv("POSTGRES_DB"),
			),
		),
		Registry: gocron.New(),
		config:   config,
	}
	NewServer.RecipeRepo = recipe.NewRecipeRepo(NewServer.NewDB)
	w := workers.Worker{DB: NewServer.NewDB}
	NewServer.Registry.Add(
		gocron.Job{
			Job:    w.CreatSelectedAndViewLog,
			Ticker: time.NewTicker(time.Minute * 2),
		},
	)

	if config.Auth != nil {
		NewServer.Auth = config.Auth
	} else {
		log.Default().Println("No Auth provided")
		NewServer.Auth = auth.NewAuth([]byte("secret"))
	}

	for _, fnc := range NewServer.config.Innit {
		err := fnc(NewServer)
		fmt.Println(err)
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(s.CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		views.Index().Render(c.Request.Context(), c.Writer)
	})

	// Add the new recipe view route
	r.GET("/view/:id", func(c *gin.Context) {
		id := c.Param("id")
		recipe, apiErr := s.RecipeRepo.GetRecipeByID(id)
		if apiErr != nil {
			c.String(apiErr.Code, apiErr.Message)
			return
		}
		views.RecipePage(recipe, "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse1.mm.bing.net%2Fth%3Fid%3DOIP.GUtzz3zgkImN3_ikBYuNfgHaE8%26pid%3DApi&f=1&ipt=e9db03ac01ccf7feb502d49d09aecfb45975d8873716e6dfa2b53c69ca00cc9c&ipo=images").Render(c.Request.Context(), c.Writer)
	})

	if s.Auth != nil {
		print("Auth")
		r.POST("/login", func(c *gin.Context) {
			s.Auth.Login(c, s.NewDB)
		})
		r.POST("/signup", func(c *gin.Context) {
			s.Auth.Signup(c, s.NewDB)
		})
		r.GET("/logout", s.Auth.Logout)
	}

	for _, controller := range s.config.Controllers {
		if controller.Route == MethodGET {
			r.GET(controller.Route, controller.Function)
		}
		if controller.Method == MethodPost {
			r.GET(controller.Route, controller.Function)
		}
	}

	r.POST("/create", s.UserMiddleware, s.AddRecipe)
	r.POST("/create_ingredient", s.AddIngredient)
	r.GET("/get", s.GetAll)
	r.GET("/popular", s.GetPopular)
	r.GET("/getbyid/:id", s.GetById)
	r.PATCH("/update/:id", s.UserMiddleware, s.UpdateRecipe)
	r.DELETE("/delete/:id", s.UserMiddleware, s.DeleteRecipe)
	r.POST("/filter", s.Filter)
	r.GET("/select/:id", s.UserMiddleware, s.Select)
	r.GET("/deselect/:id", s.UserMiddleware, s.Deselect)
	r.GET("/colormode/:type", s.Colormode)

	r.GET("/creategroup", s.UserMiddleware, s.CreateGroup)
	r.GET("/recommendation", s.UserMiddleware, s.GetRecommendation)

	return r
}
