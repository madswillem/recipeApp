package server

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/database"
)

const MethodGET = "GET"
const MethodPost = "POST"

type InnitFuncs func(*Server) error
type ExtraControllers struct {
	Function   func(*gin.Context)
	Middleware func(*gin.Context)
	Route      string
	Method     string
}

type Config struct {
	Innit       []InnitFuncs
	Controllers []ExtraControllers
}

type Server struct {
	port   int
	NewDB  *sqlx.DB
	config *Config
}

func NewServer(config *Config) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:   port,
		NewDB:  database.ConnectToDB(&sqlx.Conn{}, "user=mads password=1234 database=test_unexpected_behavior sslmode=disable"),
		config: config,
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
	tmpl := template.Must(template.New("main").ParseGlob("web/templates/**/*"))
	r.SetHTMLTemplate(tmpl)
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404/index", gin.H{
			"pageTitle": "404 Page not found",
		})
	})
	r.Use(s.CORSMiddleware())

	for _, controller := range s.config.Controllers {
		if controller.Route == MethodGET {
			r.GET(controller.Route, controller.Function)
		}
		if controller.Method == MethodPost {
			r.GET(controller.Route, controller.Function)
		}
	}

	r.POST("/create", s.AddRecipe)
	r.POST("/create_ingredient", s.AddIngredient)
	r.GET("/get", s.GetAll)
	r.GET("/getbyid/:id", s.GetById)
	r.PATCH("/update/:id", s.UpdateRecipe)
	r.DELETE("/delete/:id", s.DeleteRecipe)
	r.POST("/filter", s.Filter)
	r.GET("/select/:id", s.UserMiddleware, s.Select)
	r.GET("/deselect/:id", s.UserMiddleware, s.Deselect)
	r.GET("/colormode/:type", s.Colormode)

	r.GET("/creategroup", s.UserMiddleware, s.CreateGroup)
	r.GET("/recommendation", s.UserMiddleware, s.GetRecommendation)

	return r
}
