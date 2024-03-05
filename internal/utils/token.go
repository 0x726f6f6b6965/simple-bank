package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0x726f6f6b6965/bank/internal/proto"
	"github.com/golang-jwt/jwt/v5"
)

var (
	JwtSecretKey   []byte
	ErrTokenExpire = errors.New("the token expired")
)

// GenerateNewAccessToken generates a new JWT token
func GenerateNewAccessToken(account, nonce string, expire time.Duration) (string, error) {

	// create a JWT claim
	claims := jwt.MapClaims{}

	now := time.Now()

	// assign an expiration time for the token
	claims["expire_at"] = now.Add(expire).Unix()
	// assign a data for user
	claims["user"] = account
	// assign nonce
	claims["nonce"] = nonce
	// assign a created at time
	claims["created_at"] = now.Unix()

	// create a JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// convert the JWT token into the string
	t, err := token.SignedString(JwtSecretKey)

	// if conversion is failed, return an error
	if err != nil {
		return "", err
	}
	// return the generated JWT token
	return t, nil
}

// ExtractTokenMetadata extracts JWT token metadata
func ExtractTokenMetadata(r *http.Request) (*proto.UserToken, error) {
	// verify the JWT token
	token, err := verifyToken(r)

	// if verification is failed, return an error
	if err != nil {
		return nil, err
	}

	// get a JWT claim from the JWT token
	claims, ok := token.Claims.(jwt.MapClaims)

	// check if the token is valid
	var isValid bool = ok && token.Valid

	// if the JWT token is valid, return the JWT token metadata
	if isValid {
		// set token expiration
		expires := int64(claims["expire_at"].(float64))
		// set user for the token
		account := claims["user"].(string)
		// set nonce for the token
		nonce := claims["nonce"].(string)
		// set created at for the token
		createdAt := int64(claims["created_at"].(float64))

		// return the JWT token metadata
		return &proto.UserToken{
			ExpireAt:  expires,
			Account:   account,
			Nonce:     nonce,
			CreatedAt: createdAt,
		}, nil
	}

	// return an error
	return nil, err
}

// CheckToken checks JWT token
func CheckToken(r *http.Request) (*proto.UserToken, error) {
	// get the current time
	var now int64 = time.Now().Unix()

	// extract the JWT token metadata
	claims, err := ExtractTokenMetadata(r)
	// if extraction is failed, return an error
	if err != nil {
		return nil, err
	}

	// get the expiration time
	var expires int64 = claims.ExpireAt

	// if the token is expired, return an error
	if now > expires {
		return nil, ErrTokenExpire
	}

	// return JWT claims from the JWT token
	return claims, nil
}

// verifyToken verifies JWT token
func verifyToken(r *http.Request) (*jwt.Token, error) {
	// get the token
	var tokenString string = extractToken(r)

	// parse the JWT token
	token, err := jwt.Parse(tokenString, jwtKeyFunc)

	// if parsing is failed, return an error
	if err != nil {
		return nil, err
	}

	// return JWT token
	return token, nil
}

// extractToken extracts JWT token from the Authorization header
func extractToken(r *http.Request) string {
	// get the Authorization header
	var header string = r.Header.Get("Authorization")
	// split the content inside the header to get the JWT token
	token := strings.Split(header, " ")

	// check if the JWT token is empty
	var isEmpty bool = header == "" || len(token) < 2

	// if the JWT token is empty return an empty string
	if isEmpty {
		return ""
	}

	// return JWT token from the header
	return token[1]
}

// jwtKeyFunc return JWT secret key
func jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return JwtSecretKey, nil
}
