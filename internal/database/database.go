package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/stollenaar/ollamabot/internal/util"

	_ "github.com/marcboeker/go-duckdb/v2" // DuckDB Go driver
)

var (
	duckdbClient *sql.DB
)

func Exit() {
	duckdbClient.Close()
}

// Transaction represents a transaction record in the database.
// @name Transaction
type Transaction struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	PlatformID string    `json:"platform_id"`
	Amount     int       `json:"amount"`
	Date       time.Time `json:"date"`
	Status     string    `json:"status"`
}

// Represent a model type and cost for generation use
type Model struct {
	Name string
	Cost int
}

func init() {

	var err error

	duckdbClient, err = sql.Open("duckdb", fmt.Sprintf("%s/ollamabot.db", util.ConfigFile.DUCKDB_PATH))

	if err != nil {
		log.Fatal(err)
	}

	_, err = duckdbClient.Exec(`
		CREATE TABLE IF NOT EXISTS platforms (
			id VARCHAR,
			name VARCHAR,
			PRIMARY KEY (id)
		);
	`)

	if err != nil {
		log.Fatalf("Failed to create platforms table: %v", err)
	}

	_, err = duckdbClient.Exec(`
		CREATE TABLE IF NOT EXISTS models (
			name VARCHAR,
			cost INTEGER,
			PRIMARY KEY (name),
		);
	`)

	if err != nil {
		log.Fatalf("Failed to generate models table: %v", err)
	}

	_, err = duckdbClient.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
            id UUID DEFAULT uuid(),
			user_id VARCHAR,
			platform_id VARCHAR,
			model_name VARCHAR,
			amount INTEGER,
			date TIMESTAMP,
			status VARCHAR,
			PRIMARY KEY (id),
			FOREIGN KEY (platform_id) REFERENCES platforms(id),
			FOREIGN KEY (model_name) REFERENCES models(name)
		);
	`)

	if err != nil {
		log.Fatalf("Failed to create transactions table: %v", err)
	}
}

// AddTransaction inserts a new transaction into the database.
func AddTransaction(tx Transaction) error {
	_, err := duckdbClient.Exec(`
        INSERT INTO transactions (user_id, platform_id, amount, date, status)
        VALUES (?, ?, ?, ?, ?)
    `, tx.UserID, tx.PlatformID, tx.Amount, tx.Date, tx.Status)
	return err
}

// GetTransactionByID checks if a transaction exists and returns it by ID.
func GetTransactionByID(id string) (*Transaction, error) {
	row := duckdbClient.QueryRow(`
        SELECT id, user_id, platform_id, amount, date, status
        FROM transactions
        WHERE id = ?
    `, id)

	var tx Transaction
	err := row.Scan(&tx.ID, &tx.UserID, &tx.PlatformID, &tx.Amount, &tx.Date, &tx.Status)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetTransactionsByPlatformID checks if a transaction exists and returns it by ID.
func GetTransactionByPlatformID(id string) (transactions []*Transaction, err error) {
	rows, err := duckdbClient.Query(`
        SELECT id, user_id, platform_id, amount, date, status
        FROM transactions
        WHERE platform_id = ?
    `, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, user_id, platform_id, status string
		var amount int
		var date time.Time

		err = rows.Scan(&id, &user_id, &platform_id, &amount, &date, &status)
		if err != nil {
			break
		}

		transactions = append(transactions, &Transaction{
			ID:         id,
			UserID:     user_id,
			PlatformID: platform_id,
			Amount:     amount,
			Date:       date,
			Status:     status,
		})
	}
	return
}

// UpdateTransaction updates an existing transaction by ID.
func UpdateTransaction(tx Transaction) error {
	_, err := duckdbClient.Exec(`
        UPDATE transactions
        SET user_id = ?, platform_id = ?, amount = ?, date = ?, status = ?
        WHERE id = ?
    `, tx.UserID, tx.PlatformID, tx.Amount, tx.Date, tx.Status, tx.ID)
	return err
}

func ListModels() (models []Model, err error) {
	rows, err := duckdbClient.Query(`
	 SELECT name, cost FROM models;
	`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var name string
		var cost int

		err = rows.Scan(&name, &cost)
		if err != nil {
			break
		}

		models = append(models, Model{
			Name: name,
			Cost: cost,
		})
	}
	return
}
