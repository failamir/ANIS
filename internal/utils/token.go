package utils

import (
	"time"

	"auth-service/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func GenerateToken(userID uint, cfg *config.Config) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix() // 15 minutes
	td.AccessUuid = "access-" + time.Now().String()        // In prod use proper UUID

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix() // 7 days
	td.RefreshUuid = "refresh-" + time.Now().String()        // In prod use proper UUID

	// Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	var err error
	td.AccessToken, err = at.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(cfg.RefreshSecret))
	if err != nil {
		return nil, err
	}

	return td, nil
}
