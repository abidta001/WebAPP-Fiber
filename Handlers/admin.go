package handlers

import (
	"log"
	"regexp"

	db "myapp/DB"
	middleware "myapp/Middleware"
	models "myapp/Models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AdminResponse struct {
	Name    string
	Users   []models.UserDetails
	Invalid models.InvalidErr
}

var errors models.InvalidErr


func AdminHome(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache,no-store,must-revalidate")
	c.Set("Expires", "0")

	ok := middleware.ValidateCookie(c)
	if !ok {
		return c.Render("login", fiber.Map{
			"EmailError":    nil,
			"PasswordError": nil,
		})
	}

	role, name, err := middleware.FindRole(c)
	if err != nil {
		log.Println("Error finding role:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if role != "admin" {
		return c.Redirect("/", fiber.StatusFound)
	}

	var users []models.UserDetails
	if err := db.Db.Raw("SELECT user_name, email FROM users").Scan(&users).Error; err != nil {
		log.Println("Error fetching users:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	result := AdminResponse{
		Name:    name,
		Users:   users,
		Invalid: errors,
	}

	if err := c.Render("admin", fiber.Map{
		"title": result,
	}); err != nil {
		log.Println("Error rendering admin page:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return nil

}

func AdminAddUser(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Expires", "0")

	ok := middleware.ValidateCookie(c)
	role, _, _ := middleware.FindRole(c)
	if !ok || role != "admin" {
		return c.Redirect("/", fiber.StatusOK)
	}

	userName := c.FormValue("Name")
	userEmail := c.FormValue("Email")
	userPassword := c.FormValue("Password")

	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(userEmail) {
		errors.EmailError = "Email not in the correct format"
		return renderAdminPageWithErrors(c, errors)
	}

	var count int
	if err := db.Db.Raw("SELECT COUNT(*) FROM users WHERE email = ?", userEmail).Scan(&count).Error; err != nil {
		log.Println("Error checking user existence:", err)
		return renderAdminPageWithErrors(c, errors)
	}
	if count > 0 {
		errors.Err = "User already exists"
		return renderAdminPageWithErrors(c, errors)
	}

	userRole := "user"
	if c.FormValue("checkbox") == "on" {
		userRole = "admin"
	}

	hashedPassword, err := HashPassword(userPassword)
	if err != nil {
		log.Println("Error hashing password:", err)
		return renderAdminPageWithErrors(c, errors)
	}

	if err := db.Db.Exec("INSERT INTO users (user_name, email, password, role) VALUES(?, ?, ?, ?)", userName, userEmail, hashedPassword, userRole).Error; err != nil {
		log.Println("Error inserting user:", err)
		return renderAdminPageWithErrors(c, errors)
	}

	return c.Redirect("/admin", fiber.StatusFound)
}

func AdminUpdate(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Expires", "0")

	ok := middleware.ValidateCookie(c)
	if !ok {
		return c.Redirect("/", fiber.StatusOK)
	}

	username := c.Query("Username")
	email := c.Query("Email")

	return c.Render("updateuser", fiber.Map{
		"UserName": username,
		"Email":    email,
	})
}

func AdminUpdatePost(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Expires", "0")

	ok := middleware.ValidateCookie(c)
	if !ok {
		return c.Redirect("/", fiber.StatusOK)
	}

	email := c.FormValue("Email")
	userName := c.FormValue("Name")
	newEmail := c.FormValue("NewEmail")
	newPassword := c.FormValue("NewPassword")

	if err := db.Db.Exec("UPDATE users SET user_name = ? WHERE email = ?", userName, email).Error; err != nil {
		log.Println("Error updating user name:", err)
		return c.Redirect("/admin", fiber.StatusFound)
	}

	if newEmail != "" {
		if err := db.Db.Exec("UPDATE users SET email = ? WHERE email = ?", newEmail, email).Error; err != nil {
			log.Println("Error updating email:", err)
			return c.Redirect("/admin", fiber.StatusFound)
		}
	}

	if newPassword != "" {
		hashedPassword, hashErr := HashPassword(newPassword)
		if hashErr != nil {
			log.Println("Error hashing password:", hashErr)
			return c.Redirect("/admin", fiber.StatusFound)
		}
		if err := db.Db.Exec("UPDATE users SET password = ? WHERE email = ?", hashedPassword, email).Error; err != nil {
			log.Println("Error updating password:", err)
			return c.Redirect("/admin", fiber.StatusFound)
		}
	}

	return c.Redirect("/admin", fiber.StatusFound)
}


func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func AdminDelete(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Expires", "0")

	ok := middleware.ValidateCookie(c)
	role, _, _ := middleware.FindRole(c)
	if !ok || role != "admin" {
		return c.Redirect("/", fiber.StatusOK)
	}

	email := c.Query("Email")
	if email == "" {
		log.Println("No email provided for deletion")
		return c.Redirect("/admin", fiber.StatusFound)
	}

	if err := db.Db.Exec("DELETE FROM users WHERE email = ?", email).Error; err != nil {
		log.Println("Error deleting user:", err)
		return c.Redirect("/admin", fiber.StatusFound)
	}

	return c.Redirect("/admin", fiber.StatusFound)
}

func AdminLogout(c *fiber.Ctx) error {

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	middleware.DeleteCookie(c)

	return c.Redirect("/", fiber.StatusFound)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func renderAdminPageWithErrors(c *fiber.Ctx, errors models.InvalidErr) error {
	return c.Render("admin", fiber.Map{
		"Error": errors,
	})
}
