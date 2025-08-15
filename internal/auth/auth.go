package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytesString := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(bytesString, bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userID.String(),
	})
	signedToken, err := jwtToken.SignedString([]byte(tokenSecret))
	
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	res := claims.ExpiresAt.Time.Compare(time.Now())

	if res <= 0 {
		return uuid.Nil, errors.New("expired token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers *http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return "", errors.New("no authorization header found")
	}

	splitStrings := strings.Split(authorizationHeader, " ")
	
	if len(splitStrings) < 2 {
		return "", errors.New("no bearer token found")
	}

	return splitStrings[1], nil
}

func MakeRefreshToken() (string, error) {
	bytesArr := make([]byte, 32)
	_, err := rand.Read(bytesArr)

	if err != nil {
		return "", nil
	}

	convertedString := hex.EncodeToString(bytesArr)

	return convertedString, nil
}