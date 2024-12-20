package middleware

import (
	"errors"
	"fmt"
	"src/l"
	"src/pkg/conf"
	"time"

	"github.com/golang-jwt/jwt"
)

type SignedDetails struct {
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"firstname,omitempty"`
	LastName   string `json:"lastname,omitempty"`
	Uid        string `json:"uid,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Role       any    `json:"role,omitempty"`
	MerchantID string `json:"merchant_id,omitempty"`
	jwt.StandardClaims
}

func GenerateTokens(app *conf.Config, data SignedDetails) (token string, refreshToken string, err error) {
	claims := SignedDetails{
		Email:      data.Email,
		FirstName:  data.FirstName,
		LastName:   data.LastName,
		Uid:        data.Uid,
		Phone:      data.Phone,
		Role:       data.Role,
		MerchantID: data.MerchantID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * app.TokenLifetime).Unix(),
		},
	}
	refreshClaims := SignedDetails{
		Email:      data.Email,
		FirstName:  data.FirstName,
		LastName:   data.LastName,
		Uid:        data.Uid,
		Phone:      data.Phone,
		Role:       data.Role,
		MerchantID: data.MerchantID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(72)).Unix(),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(app.Env.SecretJWT))
	if err != nil {
		return "", "", err
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(app.Env.SecretJWT))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func ValidateToken(app *conf.Config, signedToken string) (claims *SignedDetails, err error) {
	// l.InfoF("token %s", signedToken)
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(app.Env.SecretJWT), nil
		},
	)
	if err != nil {
		l.Debug(err.Error())
		return nil, errors.New("the token in invalid")
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		return nil, errors.New("the token in invalid")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		minus_time := time.Now().Local().Unix() - claims.ExpiresAt
		msg := fmt.Sprintf("Expired Token %d", minus_time)
		return nil, errors.New(msg)
	}
	return claims, nil
}
