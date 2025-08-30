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

// Transaction represents a transaction record in the database
// @name Transaction
type Transaction struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	PlatformID string    `json:"platform_id"`
	Amount     int       `json:"amount"`
	Date       time.Time `json:"date"`
	Status     string    `json:"status"`
}

type ModelCost struct {
	PlatformName string
	Cost         int
}

// Platform represents a platform with buying power
type Platform struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BuyingPower int    `json:"buying_power"`
}

// PlatformModel represents the cost of a model on a specific platform
type PlatformModel struct {
	PlatformID string `json:"platform_id"`
	ModelName  string `json:"model_name"`
	Cost       int    `json:"cost"`
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
			buying_power INTEGER,
			PRIMARY KEY (id)
		);
	`)

	if err != nil {
		log.Fatalf("Failed to create platforms table: %v", err)
	}

	_, err = duckdbClient.Exec(`
		CREATE TABLE IF NOT EXISTS models (
			name VARCHAR,
			PRIMARY KEY (name),
		);
	`)

	if err != nil {
		log.Fatalf("Failed to generate models table: %v", err)
	}

	_, err = duckdbClient.Exec(`
        CREATE TABLE IF NOT EXISTS platform_models (
            platform_id VARCHAR,
            model_name VARCHAR,
            cost INTEGER,
            PRIMARY KEY (platform_id, model_name),
            FOREIGN KEY (platform_id) REFERENCES platforms(id),
            FOREIGN KEY (model_name) REFERENCES models(name)
        );
    `)

	if err != nil {
		log.Fatalf("Failed to create platform_models table: %v", err)
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
        VALUES (?, ?, ?, ?, ?);
    `, tx.UserID, tx.PlatformID, tx.Amount, tx.Date, tx.Status)
	return err
}

// GetTransactionByID checks if a transaction exists and returns it by ID.
func GetTransactionByID(id string) (*Transaction, error) {
	row := duckdbClient.QueryRow(`
        SELECT id, user_id, platform_id, amount, date, status
        FROM transactions
        WHERE id = ?;
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
        WHERE platform_id = ?;
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
        WHERE id = ?;
    `, tx.UserID, tx.PlatformID, tx.Amount, tx.Date, tx.Status, tx.ID)
	return err
}

// ListPlatforms lists current supported platforms
func ListPlatforms() (platforms []Platform, err error) {
	rows, err := duckdbClient.Query(`SELECT * FROM platforms;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, name string
		var buying_power int

		err = rows.Scan(&id, &name, &buying_power)
		if err != nil {
			break
		}
		platforms = append(platforms, Platform{
			ID:          id,
			Name:        name,
			BuyingPower: buying_power,
		})
	}
	return
}

// AddPlatform adds a platform
func AddPlatform(platform Platform) error {
	_, err := duckdbClient.Exec(`
		INSERT INTO platforms (id, name, buying_power)
		VALUES (?, ?, ?);
	`, platform.ID, platform.Name, platform.BuyingPower)
	return err
}

// RemovePlatform remove a platform by id
func RemovePlatform(id string) error {
	_, err := duckdbClient.Exec(`DELETE FROM platforms WHERE id = ?;`, id)
	return err
}

// GetPlatformModelCost returns the cost of a model for a specific platform.
func GetPlatformModelCost(platformID, modelName string) (int, error) {
	row := duckdbClient.QueryRow(`
        SELECT cost FROM platform_models
        WHERE platform_id = ? AND model_name = ?;
    `, platformID, modelName)

	var cost int
	err := row.Scan(&cost)
	return cost, err
}

// GetPlatformBuyingPower returns the buying power for a platform.
func GetPlatformBuyingPower(platformID string) (int, error) {
	row := duckdbClient.QueryRow(`
        SELECT buying_power FROM platforms
        WHERE id = ?;
    `, platformID)

	var buyingPower int
	err := row.Scan(&buyingPower)
	return buyingPower, err
}

func ListPlatformModels() (models map[string][]ModelCost, err error) {
	models = make(map[string][]ModelCost)

	rows, err := duckdbClient.Query(`
	 SELECT pm.model_name, pm.cost, p.name AS platform_name
	 FROM platform_models AS pm
	 JOIN platforms AS p
	 ON pm.platform_id = p.id;
	`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var platform_name, model_name string
		var cost int

		err = rows.Scan(&platform_name, &model_name, &cost)
		models[model_name] = append(models[model_name], ModelCost{
			PlatformName: platform_name,
			Cost:         cost,
		})
		if err != nil {
			break
		}

	}
	return
}

// AddModel adds a model
func AddModel(model string) error {
	_, err := duckdbClient.Exec(`
		INSERT INTO models (name)
		VALUES (?);
	`, model)
	return err
}

// ListModels lists current added models
func ListModels() (models []string, err error) {
	rows, err := duckdbClient.Query(`SELECT * FROM models;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			break
		}
		models = append(models, name)
	}
	return
}

// RemoveModel remove a model by id
func RemoveModel(id string) error {
	_, err := duckdbClient.Exec(`DELETE FROM models WHERE id = ?;`, id)
	return err
}