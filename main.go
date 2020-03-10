package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/choestelus/go-epic/db"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DB   db.Config `required:"true"`
	Port int       `required:"true" envconfig:"port"`
}

func main() {
	cfg := &Config{}
	envconfig.MustProcess("PROBE", cfg)

	app := fiber.New()
	app.Use(middleware.Recover(RecoverHandler))

	app.Get("/", HealthcheckHandler)
	app.Get("/db", DBCheckHandlerFunc(cfg.DB))

	logrus.Fatal(app.Listen(cfg.Port))
}

func HealthcheckHandler(c *fiber.Ctx) {
	c.Status(http.StatusOK).JSON(fiber.Map{"code": "ok"})
}

func DBCheckHandlerFunc(dbcfg db.Config) func(*fiber.Ctx) {
	return func(c *fiber.Ctx) {
		err := CheckDB(dbcfg)
		if err != nil {
			c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"code":    "error",
				"message": err.Error(),
			})
		} else {
			c.Status(http.StatusOK).JSON(fiber.Map{
				"code":    "ok",
				"message": "db is connected",
			})
		}
	}
}

// CheckDB re-establish connection every time when it got called
func CheckDB(dbcfg db.Config) error {
	connCfg, err := db.NewPGXConfig(dbcfg)
	if err != nil {
		logrus.Panic(err)
	}
	conn, err := pgx.ConnectConfig(context.Background(), connCfg)
	if err != nil {
		logrus.Panic(err)
	}
	defer conn.Close(context.Background())

	res := 0
	err = conn.QueryRow(context.Background(), "SELECT 1+1").Scan(&res)
	if err != nil {
		return fmt.Errorf("failed to do checking query: %w", err)
	}
	if res != 2 {
		return fmt.Errorf("got wrong query result: %v", res)
	}
	return nil
}

func RecoverHandler(c *fiber.Ctx, err error) {
	logrus.Warnf("recovered: %v", err)
	c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"message": "something went wrong",
	})
}
