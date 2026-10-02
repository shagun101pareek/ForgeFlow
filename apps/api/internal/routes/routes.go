package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shagun101pareek/forgeflow/internal/auth"
	"github.com/shagun101pareek/forgeflow/internal/generation"
	"github.com/shagun101pareek/forgeflow/internal/projects"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

type Dependencies struct {
	Log       zerolog.Logger
	DB        *pgxpool.Pool
	Redis     *redis.Client
	AppEnv    string
	JWTSecret string
	Generator generation.GenerationProvider
	DBReady   bool
	RedisOK   bool
}

func Register(app *fiber.App, deps Dependencies) {
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowHeaders:  "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders: "Content-Disposition",
		AllowMethods:  "GET,POST,PUT,PATCH,DELETE,OPTIONS",
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

	var queries *db.Queries
	if deps.DB != nil {
		queries = db.New(deps.DB)
	}

	authHandler := auth.NewHandler(queries, deps.JWTSecret)
	v1.Post("/auth/signup", authHandler.Signup)
	v1.Post("/auth/login", authHandler.Login)
	v1.Get("/auth/me", auth.RequireAuth(deps.JWTSecret), authHandler.Me)

	projectHandler := projects.NewHandler(queries)
	generationHandler := generation.NewHandler(queries, deps.Generator)
	protected := v1.Group("/projects", auth.RequireAuth(deps.JWTSecret))
	protected.Get("/", projectHandler.List)
	protected.Post("/", projectHandler.Create)
	protected.Post("/:id/generations", generationHandler.Save)
	protected.Get("/:id/generations/latest", generationHandler.Latest)
	protected.Post("/:id/generations/:generationId/github", generationHandler.Publish)
	protected.Get("/:id/generations/:generationId/export", generationHandler.Export)
	protected.Get("/:id/generations/:generationId", generationHandler.Get)
	protected.Get("/:id/generations", generationHandler.List)
	protected.Get("/:id", projectHandler.Get)
	protected.Patch("/:id", projectHandler.Update)
	protected.Delete("/:id", projectHandler.Delete)

	v1.Post("/generation", auth.RequireAuth(deps.JWTSecret), generationHandler.Create)
}
