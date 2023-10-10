package crypto

import (
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type Claims struct {
	jwt.RegisteredClaims
	Foo string `json:"foo"`
}

func TestNewJWT(t *testing.T) {
	secret := GenerateRandomBytes(64)
	tokenString, err := NewJWT(secret, jwt.RegisteredClaims{
		Audience: jwt.ClaimStrings{"test"},
		Issuer:   "example.org",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		Subject:  "cozy.io",
	})
	assert.NoError(t, err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		assert.True(t, ok, "The signing method should be HMAC")
		return secret, nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok, "Claims can be parsed as standard claims")
	assert.Equal(t, []interface{}{"test"}, claims["aud"])
	assert.Equal(t, "example.org", claims["iss"])
	assert.Equal(t, "cozy.io", claims["sub"])
}

func TestParseJWT(t *testing.T) {
	secret := GenerateRandomBytes(64)
	tokenString, err := NewJWT(secret, Claims{
		jwt.RegisteredClaims{
			Audience: jwt.ClaimStrings{"test"},
			Issuer:   "example.org",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject:  "cozy.io",
		},
		"bar",
	})
	assert.NoError(t, err)

	claims := Claims{}
	err = ParseJWT(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, &claims)
	assert.NoError(t, err)
	assert.Equal(t, jwt.ClaimStrings{"test"}, claims.Audience)
	assert.Equal(t, "example.org", claims.Issuer)
	assert.Equal(t, "cozy.io", claims.Subject)
	assert.Equal(t, "bar", claims.Foo)
}

func TestParseInvalidJWT(t *testing.T) {
	secret := GenerateRandomBytes(64)
	tokenString, err := NewJWT(secret, Claims{
		jwt.RegisteredClaims{
			Audience: jwt.ClaimStrings{"test"},
			Issuer:   "example.org",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject:  "cozy.io",
		},
		"bar",
	})
	assert.NoError(t, err)

	err = ParseJWT("invalid-token", func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, &Claims{})
	assert.Error(t, err)

	invalidSecret := GenerateRandomBytes(64)
	err = ParseJWT(tokenString, func(token *jwt.Token) (interface{}, error) {
		return invalidSecret, nil
	}, &Claims{})
	assert.Error(t, err)
}
