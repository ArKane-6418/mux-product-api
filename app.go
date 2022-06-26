package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Define struct for the app

type App struct {
	Router *mux.Router // pointer to Router
	DB     *sql.DB
}

func (a *App) initialize(user, password, dbname string) {
	// Set up connection string to use when we open the db
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable", user, password, dbname, os.Getenv("PORT"))

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func (a *App) run(addr string) {
	log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func respondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(response)
}
func respondWithError(writer http.ResponseWriter, code int, message string) {
	respondWithJSON(writer, code, map[string]string{"error": message})
}

func (a *App) getProduct(writer http.ResponseWriter, reader *http.Request) {
	vars := mux.Vars(reader)

	// Retrieve the id of the product to be fetched from the requested URL
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			// Product is not found
			respondWithError(writer, http.StatusNotFound, "Product not found")
		default:
			respondWithError(writer, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(writer, http.StatusOK, p)
}

func (a *App) getProducts(writer http.ResponseWriter, reader *http.Request) {
	count, _ := strconv.Atoi(reader.FormValue("count"))
	start, _ := strconv.Atoi(reader.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(writer, http.StatusOK, products)
}

func (a *App) createProduct(writer http.ResponseWriter, reader *http.Request) {
	var p product
	decoder := json.NewDecoder(reader.Body)

	if err := decoder.Decode(&p); err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer reader.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(writer, http.StatusCreated, p)
}

func (a *App) updateProduct(writer http.ResponseWriter, reader *http.Request) {
	vars := mux.Vars(reader)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(reader.Body)

	if err := decoder.Decode(&p); err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer reader.Body.Close()
	// Update the id
	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(writer, http.StatusOK, p)
}

func (a *App) deleteProduct(writer http.ResponseWriter, reader *http.Request) {
	vars := mux.Vars(reader)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(writer, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}
