package jwts

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(
	email string,
	uid int32,
	appID string,
	ttl time.Duration,
	secretKey string,
) (string, error) {
	const op = "internal.lib.jwt.CreateToken"

	claims := jwt.MapClaims{
		"email":  email,
		"uid":    uid,
		"app_id": appID,
		"exp":    time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}
