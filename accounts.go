package main

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	"github.com/gabivlj/jwt-api/mongodb"
	"github.com/gabivlj/jwt-api/utils"
	"golang.org/x/crypto/bcrypt"
)

/*
JWT claims struct
*/
type Token struct {
	UserID   string
	Username string
	jwt.StandardClaims
}

//a struct to rep user account
type Account struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// Create ...
func (account *Account) Create() map[string]interface{} {
	collections := mongodb.GetCollection("users")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	insertRes, err := collections.InsertOne(context.TODO(), account)
	if err != nil {
		response := utils.Message(false, "Error creating account.")
		return response
	}

	str, err := mongodb.PassObjectIDToString(insertRes.InsertedID)
	if err != nil {
		response := utils.Message(false, err.Error())
		return response
	}
	tk := Token{UserID: str, Username: account.Email}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte("token-password"))
	account.Token = tokenString
	account.Password = ""
	response := utils.Message(true, "Created account.")
	response["account"] = account

	return response
}

// LogIn ...
func (account *Account) LogIn() map[string]interface{} {
	collections := mongodb.GetCollection("users")
	user := collections.FindOne(context.TODO(), map[string]string{"email": account.Email})
	userJSON := &Account{}
	err := user.Decode(userJSON)
	if err != nil {
		response := utils.Message(false, "Error parsing account.")
		return response
	}
	err = bcrypt.CompareHashAndPassword([]byte(userJSON.Password), []byte(account.Password))
	if err != nil {
		response := utils.Message(false, "Error with the password account.")
		return response
	}
	tk := Token{Username: account.Email}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenStr, _ := token.SignedString([]byte("token-password"))
	account.Token = tokenStr
	account.Password = ""
	response := utils.Message(true, "Loged succesfuly")
	response["account"] = account
	return response
}
