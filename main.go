package main

import (
	"fmt"
	"net/http"
	"os"

	goshopify "github.com/bold-commerce/go-shopify"
	"github.com/joho/godotenv"
)

// Create an app somewhere.
var app goshopify.App

var tokens map[string]string

var app_name string

func main() {
	godotenv.Load()
	app = goshopify.App{
		ApiKey:      os.Getenv("API_KEY"),
		ApiSecret:   os.Getenv("API_SECRET"),
		RedirectUrl: os.Getenv("REDIRECT_URL"),
		Scope:       "read_products,read_orders",
	}
	tokens = make(map[string]string)
	app_name = os.Getenv("APP_NAME")

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("____________/________________________")
		fmt.Println(r.URL)
		fmt.Println(r.URL.Query())

		fmt.Println("Headers : ")
		for k, v := range r.Header {
			fmt.Println(k, " = ", v)
		}

		shop := r.URL.Query().Get("shop")
		token, exists := tokens[shop]
		if !exists {
			url := "/auth?shop=" + shop
			fmt.Println("Redirecting to", url, "...")
			http.Redirect(rw, r, url, http.StatusFound)
			return
		}
		fmt.Fprintf(rw, "Hello %s, your Oauth token is %s", shop, token)
	})

	http.HandleFunc("/auth", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("____________/auth________________________")
		fmt.Println(r.URL)
		fmt.Println(r.URL.Query())

		shopName := r.URL.Query().Get("shop")
		state := "nonce"
		authUrl := app.AuthorizeUrl(shopName, state)
		fmt.Println("Redirecting to", authUrl, "...")
		http.Redirect(rw, r, authUrl, http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("_________/auth/callback___________________________")
		fmt.Println(r.URL)
		fmt.Println(r.URL.Query())

		// Check that the callback signature is valid
		if ok, _ := app.VerifyAuthorizationURL(r.URL); !ok {
			http.Error(rw, "Invalid Signature", http.StatusUnauthorized)
			return
		}

		query := r.URL.Query()
		shopName := query.Get("shop")
		code := query.Get("code")
		token, err := app.GetAccessToken(shopName, code)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Shop (%s) has been authorized.\n", shopName)
		tokens[shopName] = token
		// url := "https://" + shopName + "/admin/apps/" + app_name

		http.Redirect(rw, r, "/", http.StatusFound)
	})

	http.ListenAndServe(":8081", nil)
}
