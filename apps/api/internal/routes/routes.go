package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Dependencies struct {
	Log     zerolog.Logger
	DB      *pgxpool.Pool
	Redis   *redis.Client
	AppEnv  string
	DBReady bool
	RedisOK bool
}

func Register(app *fiber.App, deps Dependencies) {
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		status := fiber.Map{
			"status":   "ok",
			"service":  "forgeflow-api",
			"env":      deps.AppEnv,
			"postgres": deps.DBReady,
			"redis":    deps.RedisOK,
		}

		httpStatus := fiber.StatusOK
		if !deps.DBReady || !deps.RedisOK {
			status["status"] = "degraded"
			httpStatus = fiber.StatusServiceUnavailable
		}

		return c.Status(httpStatus).JSON(status)
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		if !deps.DBReady || !deps.RedisOK {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ready": false,
			})
		}
		return c.JSON(fiber.Map{"ready": true})
	})

	v1 := app.Group("/api/v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "ForgeFlow API",
			"version": "v1",
		})
	})
}
