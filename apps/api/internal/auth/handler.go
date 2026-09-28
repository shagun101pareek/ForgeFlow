package auth

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shagun101pareek/forgeflow/internal/respond"
	"github.com/shagun101pareek/forgeflow/sql/generated"
)

type Handler struct {
	queries   *db.Queries
	jwtSecret string
}

func NewHandler(queries *db.Queries, jwtSecret string) *Handler {
	return &Handler{queries: queries, jwtSecret: jwtSecret}
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

func (h *Handler) Signup(c *fiber.Ctx) error {
	if h.queries == nil {
		return respond.Error(c, fiber.StatusServiceUnavailable, "database unavailable")
	}

	var body credentials
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	email, err := normalizeEmail(body.Email)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid email")
	}
	if err := validatePassword(body.Password); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, err.Error())
	}
	name, err := validateName(body.Name)
	if err != nil {
		return respond.Error(c, fiber.StatusBadRequest, err.Error())
	}

	hash, err := HashPassword(body.Password)
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not create account")
	}

	user, err := h.queries.CreateUser(c.Context(), db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return respond.Error(c, fiber.StatusConflict, "email already registered")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not create account")
	}

	return h.writeAuth(c, user.ID, user.Email, user.Name)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	if h.queries == nil {
		return respond.Error(c, fiber.StatusServiceUnavailable, "database unavailable")
	}

	var body credentials
	if err := c.BodyParser(&body); err != nil {
		return respond.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	email, err := normalizeEmail(body.Email)
	if err != nil || validatePassword(body.Password) != nil {
		return respond.Error(c, fiber.StatusUnauthorized, "invalid email or password")
	}

	user, err := h.queries.GetUserByEmail(c.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, _ = HashPassword(body.Password)
			return respond.Error(c, fiber.StatusUnauthorized, "invalid email or password")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not sign in")
	}

	if err := CheckPassword(user.PasswordHash, body.Password); err != nil {
		return respond.Error(c, fiber.StatusUnauthorized, "invalid email or password")
	}

	return h.writeAuth(c, user.ID, user.Email, user.Name)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	if h.queries == nil {
		return respond.Error(c, fiber.StatusServiceUnavailable, "database unavailable")
	}

	userID, ok := UserIDFrom(c)
	if !ok {
		return respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
	}

	user, err := h.queries.GetUserByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return respond.Error(c, fiber.StatusUnauthorized, "missing or invalid token")
		}
		return respond.Error(c, fiber.StatusInternalServerError, "could not load account")
	}

	return c.JSON(userResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
	})
}

func (h *Handler) writeAuth(c *fiber.Ctx, userID uuid.UUID, email, name string) error {
	token, err := IssueToken(h.jwtSecret, userID, time.Now())
	if err != nil {
		return respond.Error(c, fiber.StatusInternalServerError, "could not create session")
	}
	return c.JSON(authResponse{
		Token: token,
		User: userResponse{
			ID:    userID.String(),
			Email: email,
			Name:  name,
		},
	})
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || !strings.Contains(email, "@") {
		return "", errors.New("invalid email")
	}
	return email, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return errors.New("password must be at most 72 characters")
	}
	return nil
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	count := utf8.RuneCountInString(name)
	if count < 1 || count > 100 {
		return "", errors.New("name must be between 1 and 100 characters")
	}
	return name, nil
}
