package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FetchMonitoredWallets returns the list of wallet addresses to monitor.
// MVP: returns all addresses present in the addresses table.
// You can later scope this to a specific label (e.g., addresses where 'monitored' is in labels).
func FetchMonitoredWallets(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT address FROM addresses`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []string
	for rows.Next() {
		var addr string
		if scanErr := rows.Scan(&addr); scanErr != nil {
			return nil, scanErr
		}
		wallets = append(wallets, addr)
	}
	return wallets, rows.Err()
}
