package helpers

import (
	models "myapp/Models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("secret-key")

func CreateToken(user models.User, c *fiber.Ctx) error {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.UserName
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	signedString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    signedString,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		Secure:   true,
	})

	return nil
}

func ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		log.Fatal("Error parsing jwt", err)
	}
	return token, nil
}
