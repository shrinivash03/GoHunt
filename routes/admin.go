package routes

import (
	"ash/gohunt/db"
	"ash/gohunt/utils"
	"ash/gohunt/views"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func LoginHandler(c *fiber.Ctx) error {
	return render(c, views.Login())
}

type loginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func LoginPostHandler(c *fiber.Ctx) error {
	input := loginForm{}
	if err := c.BodyParser(&input); err != nil {
		c.Status(500)
		return c.SendString("<h2>Error: Something went wrong</h2>")
	}

	user := &db.User{}
	user, err := user.LoginAsAdmin(input.Email, input.Password)
	if err != nil {
		c.Status(401)
		return c.SendString("<h2>Error: Unauthorised access</h2>")
	}
	
	signedToken, err := utils.CreateNewAuthToken(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		c.Status(401)
		return c.SendString("<h2>Error: Something went wrong logging in</h2>")
	}
	cookie := fiber.Cookie{
		Name: "admin",
		Value: signedToken,
		Expires: time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)
	c.Append("HX-Redirect", "/")
	return c.SendStatus(200)
}

func RegisterHandler(c *fiber.Ctx) error {
	return render(c, views.Register())
}

func RegisterPostHandler(c *fiber.Ctx) error {
	input := loginForm{}
	if err := c.BodyParser(&input); err != nil {
		c.Status(500)
		return c.SendString("<h2>Error: Something went wrong</h2>")
	}
	user, err := db.CreateAdmin(input.Email, input.Password)
	if err != nil {
		c.Status(401)
		return c.SendString("<h2>Error: Unauthorised access</h2>")
	}
	
	signedToken, err := utils.CreateNewAuthToken(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		c.Status(401)
		return c.SendString("<h2>Error: Something went wrong registering in</h2>")
	}
	cookie := fiber.Cookie{
		Name: "admin",
		Value: signedToken,
		Expires: time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)
	c.Append("HX-Redirect", "/")
	return c.SendStatus(200)
}

func LogoutHandler(c *fiber.Ctx) error {
	c.ClearCookie("admin")
	c.Set("HX-Redirect", "/login")
	return c.SendStatus(200)
}

type AdminClaims struct {
	User          string `json:"user"`
	Id            string `json:"id"`
	jwt.RegisteredClaims `json:"claims"`
}

func AuthMiddleWare(c *fiber.Ctx) error {
	cookie := c.Cookies("admin")
	if cookie == "" {
		return c.Redirect("/login", 302)
	}
	token, err := jwt.ParseWithClaims(cookie, &AdminClaims{}, func(token *jwt.Token)(interface{}, error){
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return c.Redirect("/login", 302)
	}

	_, ok := token.Claims.(*AdminClaims)
	if ok && token.Valid {
		return c.Next()
	}
	return c.Redirect("/login", 302)
}

func DashboardHandler(c *fiber.Ctx) error {
	settings := &db.SearchSettings{}
	err := settings.Get()
	if err != nil {
		c.Status(500)
		return c.SendString("<h2>Error: can't get settings</h2>")
	}
	amount := strconv.FormatUint(uint64(settings.Amount), 10)
	return render(c, views.Home(amount, settings.SearchOn, settings.AddNew))
}

type settingsForm struct {
	Amount   uint   `form:"amount"`
	SearchOn string `form:"searchOn"`
	AddNew   string `form:"addNew"`
}

func DashboardPostHandler(c *fiber.Ctx) error {
	input := settingsForm{}
	if err := c.BodyParser(&input); err != nil {
		c.Status(500)
		return c.SendString("<h2>Error: can't get settings</h2>")
	}
	
	addNew := false
	if input.AddNew == "on" {
		addNew = true
	}
	
	searchOn := false
	if input.SearchOn == "on" {
		searchOn = true
	}
	
	settings := &db.SearchSettings{}
	settings.Amount = uint(input.Amount)
	settings.SearchOn = searchOn
	settings.AddNew = addNew
	err := settings.Update()
	if err != nil {
		fmt.Println(err)
		return c.SendString("<h2>Error: can't update settings</h2>")
	}
	c.Append("HX-Refresh", "true")
	return c.SendStatus(200)
}