package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Product struct {
	ID int
	Name string
	Slug string
	Price float64
}

func createProductsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS products(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		slug TEXT,
		price REAL
	);`)
	if err != nil {
		log.Fatalln("Unable to create products table", err.Error())
		panic(err)
	}
}

func addToCartHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			fmt.Fprintf(w, "Only POST method supported for this route.")
			return
		}

		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(w, "ERROR: %v", err)
			return
		}

		productId := req.FormValue("id")

		// check is there card
		ip_address := req.Header.Get("X-FORWARDED-FOR")
		var cart Cart

		err := db.QueryRow("SELECT * FROM carts WHERE ip_address = ?", ip_address).Scan(&cart.ID, &cart.IPAddress, &cart.IsCompleted)
		if err != nil {
			// create the cart
			stmt, err := db.Prepare("INSERT INTO carts(ip_address, is_completed) VALUES(?,0)")
			if err != nil {
				log.Fatal(err)
				return
			}

			res, err := stmt.Exec(ip_address)
			if err != nil {
				log.Fatal(err)
				return
			}

			id, err := res.LastInsertId()
			if err != nil {
				log.Fatal(err)
				return
			}

			cart = Cart{
				ID: int(id),
				IPAddress: ip_address,
				IsCompleted: 0,
			}
		}

		// add product to cart
		var itemId int
		var itemCount int
		err = db.QueryRow("SELECT id, product_count FROM cart_items WHERE cart_id = ? AND product_id = ?", cart.ID, productId).Scan(&itemId, &itemCount)
		if err != nil {
			stmt, err := db.Prepare("INSERT INTO cart_items(cart_id, product_id, product_count) VALUES(?,?,1)")
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = stmt.Exec(cart.ID, productId)
			if err != nil {
				log.Fatal(err)
				return
			}
		} else {
			stmt, err := db.Prepare("UPDATE cart_items set product_count = ? where id = ?")
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = stmt.Exec(itemCount + 1, itemId)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		http.Redirect(w, req, "/cart", http.StatusFound)
	}
}

func productHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		slugParameter := req.PathValue("slug")

		if slugParameter == "" {
			http.NotFound(w, req)
			return
		}

		var product Product

		err := db.QueryRow("SELECT * FROM products WHERE slug = ?", slugParameter).Scan(&product.ID, &product.Name, &product.Slug, &product.Price)
		if err != nil {
			http.NotFound(w, req)
			return
		}

		tmpl := template.Must(template.ParseFiles("./views/layout.html", "./views/product.html"))
		tmpl.Execute(w, product)
	}
}

func adminProductsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rows, err := db.Query("SELECT id, name, price FROM products")
		if err != nil {
			log.Fatal(err)
			return
		}

		var products []Product

		for rows.Next() {
			var product Product

			err = rows.Scan(&product.ID, &product.Name, &product.Price)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}

			products = append(products, product)
		}

		rows.Close()

		tmpl := template.Must(template.ParseFiles("./views/admin/products.html"))
		tmpl.Execute(w, products)
	}
}

func adminEditProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			idParameter := req.PathValue("id")

			if idParameter == "" {
				http.NotFound(w, req)
				return
			}

			var product Product

			err := db.QueryRow("SELECT * FROM products WHERE id = ?", idParameter).Scan(&product.ID, &product.Name, &product.Slug, &product.Price)
			if err != nil {
				http.NotFound(w, req)
				return
			}

			tmpl := template.Must(template.ParseFiles("./views/admin/layout.html", "./views/admin/edit-product.html"))
			tmpl.Execute(w, product)
		} else if req.Method == "POST" {
			if err := req.ParseForm(); err != nil {
				fmt.Fprintf(w, "ERROR: %v", err)
				return
			}

			id := req.FormValue("id")
			name := req.FormValue("name")
			slug := req.FormValue("slug")
			price := req.FormValue("price")

			err := db.QueryRow("SELECT id FROM products WHERE id = ?", id).Scan(&id)
			if err != nil {
				http.NotFound(w, req)
				return
			}

			stmt, err := db.Prepare("UPDATE products set name = ?, slug = ?, price = ? where id = ?")
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = stmt.Exec(name, slug, price, id)
			if err != nil {
				log.Fatal(err)
				return
			}

			http.Redirect(w, req, "/admin/products/" + id, http.StatusFound)
		} else {
			fmt.Fprintf(w, "Only GET and POST methods are supported.")
			return
		}
	}
}

func adminDeleteProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.NotFound(w, req)
			return
		}

		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(w, "ERROR: %v", err)
			return
		}

		productId := req.FormValue("id")

		stmt, err := db.Prepare("DELETE FROM products WHERE id = ?")
		if err != nil {
			log.Fatal(err)
			return
		}

		_, err = stmt.Exec(productId)
		if err != nil {
			log.Fatal(err)
			return
		}

		http.Redirect(w, req, "/admin/products", http.StatusFound)
	}
}

func adminNewProductHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			http.ServeFile(w, req, "./views/admin/new-form.html")
		} else if req.Method == "POST" {
			if err := req.ParseForm(); err != nil {
				fmt.Fprintf(w, "ERROR: %v", err)
				return
			}

			name := req.FormValue("name")
			slug := req.FormValue("slug")
			price := req.FormValue("price")

			stmt, err := db.Prepare("INSERT INTO products(name,slug,price) VALUES(?,?,?)")
			if err != nil {
				log.Fatal(err)
				return
			}

			res, err := stmt.Exec(name, slug, price)
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = res.LastInsertId()
			if err != nil {
				log.Fatal(err)
				return
			}

			http.Redirect(w, req, "/admin/products", http.StatusFound)
		} else {
			fmt.Fprintf(w, "Only GET and POST methods are supported.")
		}
	}
}
