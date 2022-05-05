package authen

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/seenark/super-backend-temp/config"

	"golang.org/x/crypto/bcrypt"
)

type TokenData struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func VerifyPassword(hashed string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func GenerateJWT(id string, email string, metamask string, role string, expire int64) (tokenData *TokenData, err error) {
	claim := jwt.MapClaims{}
	claim["id"] = id
	claim["email"] = email
	claim["metamask_address"] = metamask
	claim["role"] = role
	exp := time.Now().Add(1 * 24 * time.Hour).Unix()
	if expire > 0 {
		exp = expire
	}
	// exp := time.Now().Add(5 * time.Second).Unix()
	claim["exp"] = exp
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	cfg := config.GetConfig()
	fmt.Printf("cfg: %v\n", cfg.App.SecretKey)
	token, err := at.SignedString([]byte(cfg.App.SecretKey))
	if err != nil {
		return nil, err
	}
	return &TokenData{
		Token: token,
		Exp:   exp,
	}, nil
}

func GenerateRefreshJWT(id string, email string, jwtToken string) (*TokenData, error) {
	cfg := config.GetConfig()
	claim := jwt.MapClaims{}
	claim["id"] = id
	claim["email"] = email
	claim["jwt"] = jwtToken
	exp := time.Now().Add(7 * 24 * time.Hour).Unix()
	claim["exp"] = exp
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := at.SignedString([]byte(cfg.App.Refresh))
	if err != nil {
		return nil, err
	}
	return &TokenData{
		Token: token,
		Exp:   exp,
	}, nil
}

func ValidateJWT(jwtToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {
		cfg := config.GetConfig()
		return []byte(cfg.App.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !token.Valid || !ok {
		return claims, fmt.Errorf("jwt token invalid")
	}
	return claims, nil
}

func ValidateRefreshJWT(refreshToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		cfg := config.GetConfig()
		return []byte(cfg.App.Refresh), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("jwt token invalid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("jwt token invalid")
	}
	return claims, nil
}
