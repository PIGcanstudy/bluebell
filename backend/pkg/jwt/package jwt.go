package jwt

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGenToken(t *testing.T) {
	viper.Set("auth.jwt_expire", 2)

	// Happy path
	t.Run("HappyPath", func(t *testing.T) {
		userId := uint64(123)
		username := "testuser"
		aToken, rToken, err := GenToken(userId, username)
		assert.NoError(t, err)
		assert.NotEmpty(t, aToken)
		assert.NotEmpty(t, rToken)

		// Verify access token
		claims := &MyClaims{}
		token, err := jwt.ParseWithClaims(aToken, claims, keyFunc)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, userId, claims.UserId)
		assert.Equal(t, "username", claims.Username)
		assert.Equal(t, "bluebell", claims.Issuer)

		// Verify refresh token
		refreshClaims := &jwt.StandardClaims{}
		refreshToken, err := jwt.ParseWithClaims(rToken, refreshClaims, keyFunc)
		assert.NoError(t, err)
		assert.True(t, refreshToken.Valid)
		assert.Equal(t, "bluebell", refreshClaims.Issuer)
	})

	// Edge case: empty username
	t.Run("EmptyUsername", func(t *testing.T) {
		userId := uint64(123)
		username := ""
		aToken, rToken, err := GenToken(userId, username)
		assert.NoError(t, err)
		assert.NotEmpty(t, aToken)
		assert.NotEmpty(t, rToken)

		// Verify access token
		claims := &MyClaims{}
		token, err := jwt.ParseWithClaims(aToken, claims, keyFunc)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, userId, claims.UserId)
		assert.Equal(t, "username", claims.Username)
		assert.Equal(t, "bluebell", claims.Issuer)
	})

	// Edge case: zero userId
	t.Run("ZeroUserId", func(t *testing.T) {
		userId := uint64(0)
		username := "testuser"
		aToken, rToken, err := GenToken(userId, username)
		assert.NoError(t, err)
		assert.NotEmpty(t, aToken)
		assert.NotEmpty(t, rToken)

		// Verify access token
		claims := &MyClaims{}
		token, err := jwt.ParseWithClaims(aToken, claims, keyFunc)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, userId, claims.UserId)
		assert.Equal(t, "username", claims.Username)
		assert.Equal(t, "bluebell", claims.Issuer)
	})

	// Edge case: invalid jwt_expire config
	t.Run("InvalidJWTExpireConfig", func(t *testing.T) {
		viper.Set("auth.jwt_expire", -1)
		userId := uint64(123)
		username := "testuser"
		aToken, rToken, err := GenToken(userId, username)
		assert.NoError(t, err)
		assert.NotEmpty(t, aToken)
		assert.NotEmpty(t, rToken)

		// Verify access token
		claims := &MyClaims{}
		token, err := jwt.ParseWithClaims(aToken, claims, keyFunc)
		assert.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, userId, claims.UserId)
		assert.Equal(t, "username", claims.Username)
		assert.Equal(t, "bluebell", claims.Issuer)
	})
}
