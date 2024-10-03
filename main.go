package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	db "myapp/DB"
	handlers "myapp/Handlers"
	middleware "myapp/Middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	
	db.InitDatabase()

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use("/static", middleware.CacheControl(true))

	// User Routes
	app.Get("/", middleware.CacheControl(false), handlers.LoginPost)
	app.Post("/", middleware.CacheControl(false), handlers.LoginPost)
	app.Get("/login", middleware.CacheControl(false), handlers.LoginPost)
	app.Post("/login", middleware.CacheControl(false), handlers.LoginPost)
	app.Get("/signup", middleware.CacheControl(false), handlers.Signup)
	app.Post("/signup", middleware.CacheControl(false), handlers.SignupPost)
	app.Get("/home", middleware.CacheControl(false), handlers.Home)
	app.Get("/logout", middleware.CacheControl(false), handlers.Logout)

	// Admin Routes
	app.Get("/admin", middleware.CacheControl(false), handlers.AdminHome)
	app.Post("/admin", middleware.CacheControl(false), handlers.AdminAddUser)
	app.Get("/adminupdate", middleware.CacheControl(false), handlers.AdminUpdate)
	app.Post("/adminupdate", middleware.CacheControl(false), handlers.AdminUpdatePost)
	app.Get("/admindelete", middleware.CacheControl(false), handlers.AdminDelete)
	app.Get("/adminlogout", middleware.CacheControl(false), handlers.AdminLogout)


	go func() {
		if err := app.Listen(":8000"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()


	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	fmt.Println("Server stopped")
}
