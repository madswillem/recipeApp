package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/madswillem/recipeApp/internal/error_handler"
	"github.com/madswillem/recipeApp/internal/models"
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
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE email = $1", input.Email)
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
	user := &models.UserModel{
		Email:     input.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	// Insert into database
	query := `INSERT INTO users (email, password, created_at) VALUES ($1, $2, $3) RETURNING id`
	err = db.QueryRow(query, user.Email, user.Password, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
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

	var user models.UserModel
	err := db.Get(&user, "SELECT * FROM users WHERE email = $1", input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare password with stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (a *Auth) Verify(db *sqlx.DB, tokenString string) (*error_handler.APIError, models.UserModel) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtKey, nil
	})

	if err != nil {
		return error_handler.New("Invalid token", http.StatusUnauthorized, err), models.UserModel{}
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)

		// Get user from database
		var user models.UserModel
		err := db.Get(&user, "SELECT * FROM users WHERE id = $1", userID)
		if err != nil {
			return error_handler.New("Database error", http.StatusInternalServerError, err), models.UserModel{}
		}

		return nil, user
	}

	return error_handler.New("Invalid token", http.StatusUnauthorized, errors.New("invalid token")), models.UserModel{}
}
