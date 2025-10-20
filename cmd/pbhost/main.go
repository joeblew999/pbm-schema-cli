package main

import (
	"log"
	"os"

	pb "github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/labstack/echo/v5"
)

func main() {
	app := pb.New()
	adminEmail := os.Getenv("PB_ADMIN_EMAIL")
	adminPass  := os.Getenv("PB_ADMIN_PASSWORD")

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/healthz", func(c echo.Context) error { return c.String(200, "ok") })

		if adminEmail != "" && adminPass != "" {
			dao := app.Dao()
			if existing, _ := dao.FindAdminByEmail(adminEmail); existing == nil {
				admin := &models.Admin{Email: adminEmail}
				if err := admin.SetPassword(adminPass); err != nil {
					log.Printf("admin.SetPassword: %v", err)
				} else if err := dao.SaveAdmin(admin); err != nil {
					log.Printf("dao.SaveAdmin: %v", err)
				} else {
					log.Printf("created admin: %s", adminEmail)
				}
			}
		}
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
