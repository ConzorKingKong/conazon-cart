package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authcontrollers "github.com/conzorkingkong/conazon-users-and-auth/controllers"
)

// root should 404 - catch all non-defined routes
func TestRoot(t *testing.T) {
	t.Run("GET /", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		authcontrollers.Root(response, request)

		if response.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", response.Code, http.StatusNotFound)
		}
	})
}
