package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func homeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}

		rows, err := db.Query("SELECT * FROM products")
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		var products []Product

		for rows.Next() {
			var product Product
			err = rows.Scan(&product.ID, &product.Name, &product.Slug, &product.Price)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}

			products = append(products, product)
		}

		rows.Close()

		tmpl := template.Must(template.ParseFiles("./views/layout.html", "./views/index.html"))
		tmpl.Execute(w, products)
	}
}

func main() {
	isDbNew := false

	if _, err := os.Stat("./core.db"); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create("./core.db")
		if err != nil {
			panic(err)
		}
		f.Close()

		isDbNew = true
	}

	db, err := sql.Open("sqlite3", "./core.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if isDbNew { // run migrations
		createProjectsTable(db)
		createUsersTable(db)
		createLogsTable(db)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/product/{slug}", productHandler(db))
	mux.HandleFunc("/add-to-cart", addToCartHandler(db))
	mux.HandleFunc("/cart", cartHandler(db))
	mux.HandleFunc("/admin/products", adminProductsHandler(db))
	mux.HandleFunc("/admin/products/{id}", adminEditProductHandler(db))
	mux.HandleFunc("/admin/delete-product", adminDeleteProductHandler(db))
	mux.HandleFunc("/admin/new-product", adminNewProductHandler(db))
	mux.HandleFunc("/projects/{slug}", projectViewHandler(db))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	mux.HandleFunc("/", homeHandler(db))

	log.Println("Server started on port 8090")
	http.ListenAndServe(":8090", mux)
}
