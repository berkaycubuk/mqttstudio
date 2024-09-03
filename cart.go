package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Cart struct {
	ID int
	IPAddress string
	IsCompleted int
}

type CartItem struct {
	ID int
	CartID int
	ProductID int
	ProductCount int
}

func createCartsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS carts(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT,
		is_completed INTEGER
	);`)
	if err != nil {
		log.Fatalln("Unable to create carts table", err.Error())
		panic(err)
	}

	createCartItemsTable(db)
}

func createCartItemsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cart_items(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		cart_id INTEGER NOT NULL,
		product_id INTEGER NOT NULL,
		product_count INTEGER NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create cart items table", err.Error())
		panic(err)
	}
}

func cartHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ip_address := req.Header.Get("X-FORWARDED-FOR")

		var cart Cart

		err := db.QueryRow("SELECT * FROM carts WHERE ip_address = ?", ip_address).Scan(&cart.ID, &cart.IPAddress, &cart.IsCompleted)
		if err != nil {
			fmt.Fprintf(w, "Cart empty\n")
			return
		}

		rows, err := db.Query("SELECT * FROM cart_items WHERE cart_id = ?", cart.ID)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		var cartItems []CartItem

		for rows.Next() {
			var cartItem CartItem
			err = rows.Scan(&cartItem.ID, &cartItem.CartID, &cartItem.ProductID, &cartItem.ProductCount)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}

			cartItems = append(cartItems, cartItem)
			// TODO: should pass the product details, so it can be showed in the cart
		}

		rows.Close()

		tmpl := template.Must(template.ParseFiles("./views/layout.html", "./views/cart.html"))
		tmpl.Execute(w, cartItems)
	}
}
