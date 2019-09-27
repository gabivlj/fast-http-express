package main

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gabivlj/jwt-api/models"
	"github.com/valyala/fasthttp"
)

var JwtAuthentication = func(ctx *fasthttp.RequestCtx) error {
	tokenH := ctx.Request.Header.Peek("Authorization")
	tokenHeader := string(tokenH)
	if tokenHeader == "" {
		response := Message(false, "Unauthorized")
		ctx.SetStatusCode(500)
		RespondJSON(ctx, response)
		return nil
	}

	splitted := strings.Split(tokenHeader, " ")
	if len(splitted) != 2 {
		response := Message(false, "Invalid Token")
		ctx.SetStatusCode(500)
		RespondJSON(ctx, response)
		return nil
	}
	tokenPart := splitted[1]
	tk := &models.Token{}
	token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte("token-password"), nil
	})
	if err != nil || !token.Valid {
		response := Message(false, "Invalid Token")
		ctx.SetStatusCode(500)
		RespondJSON(ctx, response)
		return nil
	}
	fmt.Printf("User %s", tk.Username)
	AddToRequestValue(ctx, "user", tk.UserID)
	return nil

}
