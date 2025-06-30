package helpers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	authtypes "github.com/conzorkingkong/conazon-users-and-auth/types"
)

func RouteIdHelper(w http.ResponseWriter, r *http.Request) (string, int, error) {
	routeId := r.PathValue("id")

	parsedRouteId, err := strconv.Atoi(routeId)
	if err != nil {
		log.Printf("Error parsing route id: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "Internal Service Error", Data: ""})
		return "", 0, err
	}

	return routeId, parsedRouteId, nil
}