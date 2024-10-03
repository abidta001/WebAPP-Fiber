package handlers

import (
	"fmt"
	"log"
	"regexp"

	db "myapp/DB"
	helpers "myapp/Helpers"
	middleware "myapp/Middleware"
	models "myapp/Models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *fiber.Ctx) error {
	fmt.Println("Rendering Signup Page")
	return c.Render("signup", fiber.Map{
		"Error": nil,
	})
}

func SignupPost(c *fiber.Ctx) error {
	fmt.Println("Signup form submitted")
	var err models.InvalidErr
	username := c.FormValue("Name")
	email := c.FormValue("Email")
	password := c.FormValue("Password")
	confirmpassword := c.FormValue("ConfirmPassword")


	if !isValidEmail(email) {
		err.EmailError = "Email not in proper format"
		return c.Render("signup", fiber.Map{"EmailError": err.EmailError})
	}


	if password != confirmpassword {
		err.PasswordError = "Passwords do not match"
		return c.Render("signup", fiber.Map{"PasswordError": err.PasswordError})
	}


	if userExists(email) {
		err.EmailError = "User already exists"
		return c.Render("signup", fiber.Map{"EmailError": err.EmailError})
	}


	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		log.Println("Error hashing password:", hashErr)
		return c.Render("signup", fiber.Map{"Error": "Error hashing password"})
	}


	if insertErr := db.Db.Exec("INSERT INTO users(user_name, email, password) VALUES(?, ?, ?)", username, email, hashedPassword).Error; insertErr != nil {
		log.Println("Error inserting user:", insertErr)
		return c.Render("signup", fiber.Map{"Error": "Error inserting data"})
	}

	return c.Redirect("/login", fiber.StatusFound) 
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(pattern).MatchString(email)
}

func userExists(email string) bool {
	var count int
	if err := db.Db.Raw("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count).Error; err != nil {
		log.Println("Error checking user existence:", err)
		return false
	}
	return count > 0
}

func LoginPost(c *fiber.Ctx) error {
	email := c.FormValue("Email")
	password := c.FormValue("Password")

	var compare models.Compare

	fmt.Println("Posted")

	if err := db.Db.Raw("SELECT password, role, user_name FROM users WHERE email=?", email).Scan(&compare).Error; err != nil {
		log.Println("Error querying user:", err)
		return c.Render("login", fiber.Map{"EmailError": "Invalid email or password"})
	}

	if bcrypt.CompareHashAndPassword([]byte(compare.Password), []byte(password)) != nil {
		log.Println("Invalid password")
		return c.Render("login", fiber.Map{"PasswordError": "Invalid password"})
	}


	user := models.User{
		Role:     compare.Role,
		UserName: compare.UserName,
	}

	helpers.CreateToken(user, c)

	if compare.Role == "admin" {
		return c.Redirect("/admin", fiber.StatusFound)
	}
	return c.Redirect("/home", fiber.StatusFound)
}

func Home(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache, no-store")
	c.Set("Expires", "0")

	ok := middleware.ValidateCookie(c)
	role, user, _ := middleware.FindRole(c)
	if !ok || role != "user" {
		return c.Redirect("/", fiber.StatusFound)
	}

	return c.Render("home", fiber.Map{"UserName": user})
}

func Logout(c *fiber.Ctx) error {
	middleware.DeleteCookie(c)

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	return c.Redirect("/", fiber.StatusFound)
}
