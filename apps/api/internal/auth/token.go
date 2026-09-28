package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenTTL = 7 * 24 * time.Hour

type claims struct {
	jwt.RegisteredClaims
}

func IssueToken(secret string, userID uuid.UUID, now time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func ParseToken(secret, raw string) (uuid.UUID, error) {
	parsed, err := jwt.ParseWithClaims(raw, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return uuid.Nil, err
	}

	tokenClaims, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(tokenClaims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject: %w", err)
	}
	return userID, nil
}
