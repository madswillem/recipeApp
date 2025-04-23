package apierror

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Message string
	Code    int
	Errors  []error
}

func New(message string, code int, errors error) *APIError {
	return &APIError{
		Message: message,
		Code:    code,
		Errors:  []error{errors},
	}
}

func (apierr *APIError) Error() string {
	return fmt.Sprintf("API Error: %s, Code: %d, Errors: %v", apierr.Message, apierr.Code, apierr.Errors)
}

func (apierr *APIError) Add(err error) {
	apierr.Errors = append(apierr.Errors, err)
}

func (apierr *APIError) Handle(c *gin.Context) {
	log.Default().Println(apierr.Error())
	c.JSON(apierr.Code, gin.H{
		"error":   apierr.Message,
		"details": apierr.Errors,
	})
	c.Abort()
}

func HandleError(c *gin.Context, statusCode int, errorMessage string, err []error) {
	c.AbortWithStatusJSON(statusCode, gin.H{
		"errors":     err[0].Error(),
		"errMessage": errorMessage,
	})
}
