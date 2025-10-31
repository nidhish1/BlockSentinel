package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Address struct {
	Address   string     `json:"address"`
	FirstSeen *time.Time `json:"first_seen,omitempty"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	Labels    []string   `json:"labels,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func registerAddressRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	// POST /addresses
	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var in Address
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
			if strings.TrimSpace(in.Address) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "address required"})
				return
			}
			ctx := context.Background()
			_, err := db.Exec(ctx,
				`INSERT INTO addresses(address, first_seen, last_seen, labels)
                 VALUES ($1, $2, $3, $4)
                 ON CONFLICT (address) DO UPDATE SET first_seen = COALESCE(EXCLUDED.first_seen, addresses.first_seen),
                                             last_seen = COALESCE(EXCLUDED.last_seen, addresses.last_seen),
                                             labels = COALESCE(EXCLUDED.labels, addresses.labels),
                                             updated_at = NOW()`,
				in.Address, in.FirstSeen, in.LastSeen, toTextArray(in.Labels),
			)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
		case http.MethodGet:
			// Optional: list with pagination
			// For brevity, return 405 to avoid an unbounded list by default
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use /addresses/{address}"})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// GET/PUT/DELETE /addresses/{address}
	mux.HandleFunc("/addresses/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/addresses/")
		if path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "address required"})
			return
		}
		addr := path
		ctx := context.Background()

		switch r.Method {
		case http.MethodGet:
			var out Address
			var labels []string
			err := db.QueryRow(ctx,
				`SELECT address, first_seen, last_seen, labels, created_at, updated_at
                 FROM addresses WHERE address = $1`, addr,
			).Scan(&out.Address, &out.FirstSeen, &out.LastSeen, &labels, &out.CreatedAt, &out.UpdatedAt)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
				return
			}
			out.Labels = labels
			writeJSON(w, http.StatusOK, out)

		case http.MethodPut:
			var in Address
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
			_, err := db.Exec(ctx,
				`UPDATE addresses SET first_seen=$2, last_seen=$3, labels=$4, updated_at=NOW() WHERE address=$1`,
				addr, in.FirstSeen, in.LastSeen, toTextArray(in.Labels),
			)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

		case http.MethodDelete:
			_, err := db.Exec(ctx, `DELETE FROM addresses WHERE address=$1`, addr)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

// toTextArray converts a slice to a Postgres text[] compatible value.
func toTextArray(v []string) []string { return v }
