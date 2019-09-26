package main

import (
	"flag"
	"fmt"

	"github.com/valyala/fasthttp"
)

var (
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
	compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
)
var nOfIther = 0

// Simple example of how fast-http-express works
func main() {

	router1 := NewRouter("/api/go")
	router2 := NewRouter("/api/http")

	router2.Get("/32", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, "I guess it works...")
		return nil
	})

	router1.Get("/lol/:id", func(ctx *fasthttp.RequestCtx) error {
		nOfIther++
		fmt.Println(nOfIther)
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, "____yes____"+GetParams(ctx)["id"]+string(GetParams(ctx)["id"]))
		return nil
	})

	router1.Delete("/lol/:id/:the_id_especial", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, GetParams(ctx)["the_id_especial"])
		return nil
	})

	router1.Post("/lol/:id", func(ctx *fasthttp.RequestCtx) error {
		ctx.SetStatusCode(200)
		fmt.Fprintf(ctx, string(ctx.PostBody()))
		return nil
	})

	app := NewApp(*router1, *router2)
	app.StartApp(":8080")

}
