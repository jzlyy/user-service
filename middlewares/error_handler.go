package middlewares

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"user-service/database"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("[ERROR] %s %s: %+v", c.Request.Method, c.Request.URL.Path, err.Err)

			switch {
			case errors.Is(err.Err, sql.ErrNoRows):
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			case errors.Is(err.Err, database.RedisError):
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
			default:
				if gin.Mode() == gin.DebugMode {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Err.Error(),
						"stack": fmt.Sprintf("%+v", err.Err),
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				}
			}
			c.Abort()
		}
	}
}
