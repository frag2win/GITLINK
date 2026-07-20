package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Auth returns a Fiber middleware handler that validates the user's
// identity on each request using JWT.
func Auth(jwtSecret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenString := ""
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// Fallback to query string or WebSocket subprotocol headers for EventSource/WebSocket support
		if tokenString == "" {
			tokenString = c.Query("token")
		}
		if tokenString == "" {
			wsProto := c.Get("Sec-WebSocket-Protocol")
			if wsProto != "" {
				tokenString = strings.TrimSpace(wsProto)
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		userID := uint(claims["sub"].(float64))
		c.Locals("userID", userID)
		c.Locals("username", claims["username"])

		return c.Next()
	}
}

// UserIDFromContext extracts the authenticated user ID.
func UserIDFromContext(c *fiber.Ctx) uint {
	if id, ok := c.Locals("userID").(uint); ok {
		return id
	}
	return 0
}
