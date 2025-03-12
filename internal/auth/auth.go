package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/madswillem/recipeApp/internal/recipe"
	"github.com/madswillem/recipeApp/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	jwtKey []byte
}

func NewAuth(key []byte) *Auth {
	return &Auth{
		jwtKey: key,
	}
}

func (a *Auth) Signup(c *gin.Context, db *sqlx.DB) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM public.user WHERE email = $1", input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
		return
	}

	// Generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	// Create new user
	user := &user.UserModel{
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	// Insert into database
	_, err = db.Exec(`INSERT INTO public.user (email, password) VALUES ($1, $2)`, user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	c.Status(http.StatusCreated)
}

func (a *Auth) Login(c *gin.Context, db *sqlx.DB) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user user.UserModel
	err := db.Get(&user, "SELECT id, password, email FROM public.user WHERE email = $1", input.Email)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare password with stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(a.jwtKey)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating token"})
		return
	}

	c.SetCookie("token", tokenString, 60*60*24, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (a *Auth) Verify(db *sqlx.DB, tokenString string) (*error_handler.APIError, user.UserModel) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtKey, nil
	})

	if err != nil {
		return error_handler.New("Invalid token", http.StatusUnauthorized, err), user.UserModel{}
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Create user from claims
		user := user.UserModel{
			ID:    claims["user_id"].(string),
			Email: claims["email"].(string),
		}
		return nil, user
	}

	return error_handler.New("Invalid token", http.StatusUnauthorized, errors.New("invalid token")), user.UserModel{}
}

func (a *Auth) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (a *Auth) AccessControl(sub string, obj string, act string, db *sqlx.DB) (bool, *error_handler.APIError) {
	recipe := recipe.RecipeSchema{ID: obj}
	owner, ownererr := recipe.GetAuthor(db)
	if ownererr != nil {
		return false, ownererr
	}
	if owner != sub {
		return false, error_handler.New("Unauthorized", http.StatusUnauthorized, errors.New("Unauthorized"))
	}

	return true, nil
}