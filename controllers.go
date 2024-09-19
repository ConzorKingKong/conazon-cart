package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func routeIdHelper(w http.ResponseWriter, r *http.Request) (string, int, error) {
	routeId := r.PathValue("id")

	parsedRouteId, err := strconv.Atoi(routeId)
	if err != nil {
		log.Printf("Error parsing route id: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "Internal Service Error", Data: ""})
		return "", 0, err
	}

	return routeId, parsedRouteId, nil
}

func Root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "invalid path" + r.URL.RequestURI(), Data: ""})
}

func CartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		TokenData, err := validateAndReturnSession(w, r)
		if err != nil {
			return
		}

		type Call struct {
			ProductId int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		call := Call{}

		err = json.NewDecoder(r.Body).Decode(&call)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Do something with the Person struct...
		fmt.Printf("Call: %+v \n", call)

		conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		var id int

		err = conn.QueryRow(context.Background(), "insert into cart.cart (user_id, product_id, quantity, status) values ($1, $2, $3, $4) returning id", TokenData.Id, call.ProductId, call.Quantity, "active").Scan(&id)
		if err != nil {
			log.Printf("Error making cart with user id %d - %s", TokenData.Id, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "Cart could not be made", Data: ""})
			return
		}

		// return data
		json.NewEncoder(w).Encode(CartResponse{Status: http.StatusOK, Message: "Success", Data: Cart{ID: id, UserID: TokenData.Id, ProductID: call.ProductId, Quantity: call.Quantity, Status: "active"}})
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Status: http.StatusMethodNotAllowed, Message: "Method Not Allowed", Data: ""})
		return
	}
}

func CartId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	routeId, _, err := routeIdHelper(w, r)
	if err != nil {
		return
	}

	if r.Method == "GET" {
		TokenData, err := validateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		cart := Cart{}

		err = conn.QueryRow(context.Background(), "select id, user_id, product_id, quantity, status from cart.cart where id=$1", routeId).Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		if TokenData.Id != cart.UserID {
			log.Printf("Error: user tried reading cart they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		// return data
		// json.NewEncoder(w).Encode(cart)
		json.NewEncoder(w).Encode(CartResponse{Status: http.StatusOK, Message: "Success", Data: cart})

	} else if r.Method == "PATCH" {
		// update quantity of product in cart
		TokenData, err := validateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		cart := Cart{}

		// verify owner of cart with db call
		err = conn.QueryRow(context.Background(), "select user_id, status from cart.cart where id=$1", routeId).Scan(&cart.UserID, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		if TokenData.Id != cart.UserID {
			log.Printf("Error: user tried reading cart they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		if cart.Status != "active" {
			log.Printf("Error: do not edit deleted or completed cart")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		type Call struct {
			Quantity int `json:"quantity"`
		}

		var call Call

		json.NewDecoder(r.Body).Decode(&call)

		// call becomes change quantity
		_, err = conn.Exec(context.Background(), "update cart.cart set quantity = $1 where id=$2", call.Quantity, routeId)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		err = conn.QueryRow(context.Background(), "select id, user_id, product_id, quantity, status from cart.cart where id=$1", routeId).Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		// return this cart or the whole cart???
		// json.NewEncoder(w).Encode(cart)
		json.NewEncoder(w).Encode(CartResponse{Status: http.StatusOK, Message: "Success", Data: cart})
	} else if r.Method == "DELETE" {
		// set cart status to deleted
		TokenData, err := validateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		cart := Cart{}

		// verify owner of cart with db call
		err = conn.QueryRow(context.Background(), "select user_id, status from cart.cart where id=$1", routeId).Scan(&cart.UserID, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		if TokenData.Id != cart.UserID {
			log.Printf("Error: user tried reading cart they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		if cart.Status == "deleted" {
			log.Printf("Error: user tried deleting cart they already deleted")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "Error cart already deleted", Data: ""})
			return
		}

		// call becomes change quantity
		_, err = conn.Exec(context.Background(), "update cart.cart set status = 'deleted' where id=$1", routeId)
		if err != nil {
			log.Printf("Error deleting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		// return whole cart or nothing???
		// json.NewEncoder(w).Encode(parsedRouteId)
		json.NewEncoder(w).Encode(Response{Status: http.StatusOK, Message: "Cart Deleted", Data: ""})
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Status: http.StatusMethodNotAllowed, Message: "Method Not Allowed", Data: ""})
		return
	}
}

func UserId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	routeId, parsedRouteId, err := routeIdHelper(w, r)
	if err != nil {
		return
	}

	if r.Method == "GET" {
		// get all cart items for user
		TokenData, err := validateAndReturnSession(w, r)
		if err != nil {
			return
		}

		if TokenData.Id != parsedRouteId {
			log.Printf("Error: You are not authorized to get this users cart")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "You are not authorized to get this users cart", Data: ""})
			return
		}
		// make database call
		conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		// cart := Cart{}
		// .Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
		rows, err := conn.Query(context.Background(), "select id, user_id, product_id, quantity, status from cart.cart where user_id=$1 and status = 'active'", TokenData.Id)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		var rowSlice []Cart

		for rows.Next() {
			var cart Cart
			err = rows.Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
			if err != nil {
				log.Printf("Error getting cart with id %d - %s", TokenData.Id, err)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Error loading cart", Data: ""})
				return
			}
			rowSlice = append(rowSlice, cart)
		}

		defer rows.Close()

		if rowSlice == nil {
			json.NewEncoder(w).Encode(UserCartResponse{Status: http.StatusOK, Message: "No cart found for user", Data: []Cart{}})
			return
		}

		// return data
		// json.NewEncoder(w).Encode(rowSlice)
		json.NewEncoder(w).Encode(UserCartResponse{Status: http.StatusOK, Message: "Success", Data: rowSlice})
	} else if r.Method == "PUT" {
		// set all active cart items to purchased

		TokenData, err := validateAndReturnSession(w, r)
		if err != nil {
			return
		}

		if TokenData.Id != parsedRouteId {
			log.Printf("Error: You are not authorized to update this user")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Status: http.StatusUnauthorized, Message: "You are not authorized to update this user", Data: ""})
			return
		}
		// make database call
		conn, err := pgx.Connect(context.Background(), DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "update cart.cart set status = 'purchased' where user_id=$1 and status = 'active'", TokenData.Id)

		if err != nil {
			log.Printf("Error updating cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Status: http.StatusNotFound, Message: "Error updating cart", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(Response{Status: 200, Message: "cart purchase completed", Data: ""})

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Status: http.StatusMethodNotAllowed, Message: "Method Not Allowed", Data: ""})
		return
	}
}
