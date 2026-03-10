package Infrastructure

import (
	"net/http"
	"strings"

	domainInterface "bloodlink/Domain/Interfaces"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware extracts the JWT, verifies it using IAuthentication,
// and checks if the user's role is in the allowedRoles list.
func AuthMiddleware(auth domainInterface.IAuthentication, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Bearer schema
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse token
		claims, err := auth.ParseTokenToClaim(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Role Check
		if len(allowedRoles) > 0 {
			roleAllowed := false
			for _, allowedRole := range allowedRoles {
				if claims.AccountType == allowedRole {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this resource"})
				c.Abort()
				return
			}
		}

		// Store user information in context for down-stream handlers
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.AccountType)

		c.Next()
	}
}
