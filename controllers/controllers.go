package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/conzorkingkong/conazon-cart/config"
	"github.com/conzorkingkong/conazon-cart/token"
	"github.com/conzorkingkong/conazon-cart/types"
	authhelpers "github.com/conzorkingkong/conazon-users-and-auth/helpers"
	authtypes "github.com/conzorkingkong/conazon-users-and-auth/types"
	"github.com/jackc/pgx/v5"
)

func CartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET CURRENT USERS CARTS (carts+items that are active)
	if r.Method == "GET" {
		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "select id, user_id, product_id, quantity, status from cart.cart where user_id=$1 and status = 'active'", TokenData.Id)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", TokenData.Id, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		var rowSlice []types.Cart

		for rows.Next() {
			var cart types.Cart
			err = rows.Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
			if err != nil {
				log.Printf("Error getting cart with id %d - %s", TokenData.Id, err)
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Error loading cart", Data: ""})
				return
			}
			rowSlice = append(rowSlice, cart)
		}

		defer rows.Close()

		if rowSlice == nil {
			json.NewEncoder(w).Encode(types.UserCartResponse{Status: http.StatusOK, Message: "No cart found for user", Data: []types.Cart{}})
			return
		}

		json.NewEncoder(w).Encode(types.UserCartResponse{Status: http.StatusOK, Message: "Success", Data: rowSlice})

		// CREATE CART with items
	} else if r.Method == "POST" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		call := types.Call{}

		err = json.NewDecoder(r.Body).Decode(&call)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		var id int

		err = conn.QueryRow(context.Background(), "insert into cart.cart (user_id, product_id, quantity, status) values ($1, $2, $3, $4) returning id", TokenData.Id, call.ProductId, call.Quantity, "active").Scan(&id)
		if err != nil {
			log.Printf("Error making cart with user id %d - %s", TokenData.Id, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "Cart could not be made", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(types.CartResponse{Status: http.StatusOK, Message: "Success", Data: types.Cart{ID: id, UserID: TokenData.Id, ProductID: call.ProductId, Quantity: call.Quantity, Status: "active"}})

		// CHECKOUT PROCESS CALL TO MARK ACTIVE CARTS AS PURCHASED
	} else if r.Method == "PATCH" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "update cart.cart set status='purchased' where user_id=$1 and status='active'", TokenData.Id)

		if err != nil {
			log.Printf("Error updating cart with id %d - %s", TokenData.Id, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Error updating cart", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(authtypes.Response{Status: 200, Message: "cart status change to purchased completed", Data: ""})

		// DELETE ALL ACTIVE CART/ITEM COMBOS
	} else if r.Method == "DELETE" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		// delete cart
		_, err = conn.Exec(context.Background(), "update cart.cart set status = 'deleted' where user_id=$1 and status=$2", TokenData.Id, "active")
		if err != nil {
			log.Printf("Error deleting carts for user %d - %s", TokenData.Id, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusOK, Message: "Cart Deleted", Data: ""})

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusMethodNotAllowed, Message: "Method Not Allowed", Data: ""})
		return
	}
}

func CartId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	routeId, _, err := authhelpers.RouteIdHelper(w, r)
	if err != nil {
		return
	}

	// GET cart by id, regardless of status
	if r.Method == "GET" {
		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		cart := types.Cart{}

		err = conn.QueryRow(context.Background(), "select id, user_id, product_id, quantity, status from cart.cart where id=$1", routeId).Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		if TokenData.Id != cart.UserID {
			log.Printf("Error: user tried reading cart they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(types.CartResponse{Status: http.StatusOK, Message: "Success", Data: cart})

		// update quantity of product in active cart
	} else if r.Method == "PATCH" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		cart := types.Cart{}

		// verify owner of cart with db call
		err = conn.QueryRow(context.Background(), "select user_id, status from cart.cart where id=$1", routeId).Scan(&cart.UserID, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		if TokenData.Id != cart.UserID {
			log.Printf("Error: user tried reading cart they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		if cart.Status != "active" {
			log.Printf("Error: do not edit deleted or completed cart")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
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
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		err = conn.QueryRow(context.Background(), "select id, user_id, product_id, quantity, status from cart.cart where id=$1", routeId).Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(types.CartResponse{Status: http.StatusOK, Message: "Success", Data: cart})

		// set cart status to deleted
	} else if r.Method == "DELETE" {

		TokenData, err := token.ValidateAndReturnSession(w, r)
		if err != nil {
			return
		}

		// make database call
		conn, err := pgx.Connect(context.Background(), config.DatabaseURLEnv)
		if err != nil {
			log.Printf("Error connecting to database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
			return
		}

		defer conn.Close(context.Background())

		cart := types.Cart{}

		// verify owner of cart with db call
		err = conn.QueryRow(context.Background(), "select user_id, status from cart.cart where id=$1", routeId).Scan(&cart.UserID, &cart.Status)
		if err != nil {
			log.Printf("Error getting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		if TokenData.Id != cart.UserID {
			log.Printf("Error: user tried reading cart they don't own")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "Unauthorized", Data: ""})
			return
		}

		if cart.Status == "deleted" {
			log.Printf("Error: user tried deleting cart they already deleted")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "Error cart already deleted", Data: ""})
			return
		}

		// delete cart
		_, err = conn.Exec(context.Background(), "update cart.cart set status = 'deleted' where id=$1", routeId)
		if err != nil {
			log.Printf("Error deleting cart with id %s - %s", routeId, err)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusNotFound, Message: "Cart not found", Data: ""})
			return
		}

		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusOK, Message: "Cart Deleted", Data: ""})
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusMethodNotAllowed, Message: "Method Not Allowed", Data: ""})
		return
	}
}
