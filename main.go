package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabivlj/jwt-api/models"
	"github.com/gabivlj/jwt-api/mongodb"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var nOfIther = 0

// Simple example of how fast-http-express works
func main() {
	// MONGO
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	mongodb.SetClient(client)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Connected to mongodb!")

	router1 := NewRouter("/api/go")
	router2 := NewRouter("/api/http")
	authentification := NewRouter("/api/prohibited")
	auth := NewRouter("/api/auth")

	router2.Get("/32", func(ctx *fasthttp.RequestCtx) error {
		AddToRequestValue(ctx, "wow", "nice")
		key := RequestKeyValueBytes(ctx, "wow")
		RespondBytes(ctx, key)
		return nil
	}, func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		RespondJSON(ctx, map[string]interface{}{"lol": 32, "yeees": "owo", "user": TestUser{Username: "hehehhee"}})
		return nil
	})

	router1.Get("/lol/:id", func(ctx *fasthttp.RequestCtx) error {
		nOfIther++
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, "____yes____"+GetParams(ctx)["id"]+string(GetParams(ctx)["id"]))
		return nil
	})

	router1.Delete("/lol/:id/:the_id_especial", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, GetParams(ctx)["the_id_especial"])
		return nil
	})

	auth.Post("/log-in", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)

		account := &models.Account{}
		err := json.Unmarshal(ctx.Request.Body(), account)
		if err != nil {
			// utils.Respond(w, utils.Message(false, "Invalid request"))
			ctx.SetStatusCode(500)
			RespondJSON(ctx, Message(false, "Invalid request"))
			return err
		}

		resp := account.LogIn()
		RespondJSON(ctx, resp)
		return nil
	})

	auth.Post("/sign-in", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)

		account := &models.Account{}
		err := json.Unmarshal(ctx.Request.Body(), account)
		if err != nil {
			ctx.SetStatusCode(500)
			RespondJSON(ctx, Message(false, "Invalid request"))
			return err
		}

		resp := account.LogIn()
		RespondJSON(ctx, resp)
		return nil
	})

	authentification.Get("/", JwtAuthentication, func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		RespondText(ctx, "yex")
		return nil
	})

	router1.Post("/lol/:id", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, string(ctx.PostBody()))
		return nil
	})

	// todo: be able to add middleware to app.
	app := NewApp(*router1, *router2, *auth, *authentification)
	app.StartApp(":8080")

}
