package crypt

import (
	"os"
	"time"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"

)

var jwtSecret []byte

func init() {
	_ = godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	jwtSecret = []byte(secret)
}

func GenerateJWT(userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
}

func GetClaims(token *jwt.Token) (jwt.MapClaims, error) {
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token claims")
    }
    return claims, nil
}
func GetUserIDFromJWT(token *jwt.Token) (uint, error) {
    claims, err := GetClaims(token)
    if err != nil {
        return 0, err
    }
    uidFloat, ok := claims["user_id"].(float64)
    if !ok {
        return 0, errors.New("user_id not found in token")
    }

    return uint(uidFloat), nil
}