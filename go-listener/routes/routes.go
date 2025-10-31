package routes

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterRoutes wires all HTTP routes.
func RegisterRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	registerAddressRoutes(mux, db)
	// Add more route groups here
}
