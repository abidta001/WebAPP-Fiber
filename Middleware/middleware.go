package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func ValidateCookie(c *fiber.Ctx) bool {
	cookie := c.Cookies("jwt")
	if cookie == "" {
		fmt.Println("No cookie found")
		return false
	}

	token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte("secret"), nil
	})

	if err != nil {
		fmt.Println("Error parsing token:", err)
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expiration := int64(claims["exp"].(float64))
		if expiration < time.Now().Unix() {
			fmt.Println("Token has expired")
			return false
		}
		fmt.Println("Valid token for user:", claims["name"])
		return true
	} else {
		fmt.Println("Invalid token")
		return false
	}
}

func DeleteCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:    "jwt",
		Expires: time.Now().Add(-time.Hour),
	})
}

func FindRole(c *fiber.Ctx) (string, string, error) {
	cookie := c.Cookies("jwt")
	if cookie == "" {
		return "", "", fmt.Errorf("no cookie found")
	}

	token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte("secret"), nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		role := claims["role"].(string)
		user := claims["name"].(string)
		return role, user, nil
	}

	return "", "", fmt.Errorf("invalid token")
}


func CacheControl(cacheable bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cacheable {
			//1 hr
			c.Set("Cache-Control", "public, max-age=3600")
		} else {
			c.Set("Cache-Control", "no-store")
		}
		return c.Next()
	}
}


func CheckLogin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !isLoggedIn(c) {
			c.Locals("isLoggedIn", false)
			return c.Next()
		}
		c.Locals("isLoggedIn", true)
		return c.Next()
	}
}

func isLoggedIn(c *fiber.Ctx) bool {
	return c.Locals("user") != nil || ValidateCookie(c) 
}
