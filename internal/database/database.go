package database

import (
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/stollenaar/ollamabot/internal/util"

	_ "github.com/marcboeker/go-duckdb/v2" // DuckDB Go driver
)

var (
	duckdbClient *sql.DB

	//go:embed changelog/*.sql
	changeLogFiles embed.FS
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
	Tokens       int
}

// Platform represents a platform with buying power
type Platform struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BuyingPower int    `json:"buying_power"`
}

// PlatformModel represents the tokens of a model on a specific platform
type PlatformModel struct {
	PlatformID string `json:"platform_id"`
	ModelName  string `json:"model_name"`
	Tokens     int    `json:"tokens"`
}

// History track all prompts made with the bot
type History struct {
	ID        int    `json:"id"`
	UserID    string `json:"user_id"`
	ModelName string `json:"model_name"`
	Prompt    string `json:"prompt"`
}

// UserContext track the user contexts
type UserContext struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	ModelName string    `json:"model_name"`
	Context   []int32   `json:"context"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Thread records
type Thread struct {
	ThreadID  string
	Context   []int32
	Prompt    string
	ModelName string
}

func init() {

	var err error

	duckdbClient, err = sql.Open("duckdb", fmt.Sprintf("%s/ollamabot.db", util.ConfigFile.DUCKDB_PATH))

	if err != nil {
		log.Fatal(err)
	}

	// Ensure changelog table exists
	_, err = duckdbClient.Exec(`
	CREATE TABLE IF NOT EXISTS database_changelog (
		id INTEGER PRIMARY KEY,
		name VARCHAR NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		checksum VARCHAR,
		success BOOLEAN DEFAULT TRUE
	);
	`)

	if err != nil {
		log.Fatalf("failed to create changelog table: %v", err)
	}

	if err := runMigrations(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	slog.Info("All migrations applied successfully.")
}

func runMigrations() error {
	entries, err := changeLogFiles.ReadDir("changelog")
	if err != nil {
		return fmt.Errorf("failed to read embedded changelogs: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)

	for i, file := range files {
		id := i + 1

		contents, err := changeLogFiles.ReadFile(filepath.Join("changelog", file))
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		checksum := sha256.Sum256(contents)
		checksumHex := hex.EncodeToString(checksum[:])

		var appliedChecksum string
		err = duckdbClient.QueryRow("SELECT checksum FROM database_changelog WHERE id = ?", id).Scan(&appliedChecksum)
		if err == nil {
			if appliedChecksum != checksumHex {
				return fmt.Errorf("checksum mismatch for migration %s (id=%d). File has changed", file, id)
			}
			log.Printf("Skipping already applied migration %s", file)
			continue
		}

		// Run changelogs in a transaction
		tx, err := duckdbClient.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin tx: %w", err)
		}

		_, err = tx.Exec(string(contents))
		if err != nil {
			_ = tx.Rollback()
			_, _ = duckdbClient.Exec(`
				INSERT INTO database_changelog (id, name, applied_at, checksum, success) VALUES (?, ?, ?, ?, false)
			`, id, file, time.Now(), checksumHex)
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		_, err = duckdbClient.Exec(`
			INSERT INTO database_changelog (id, name, applied_at, checksum, success)
			VALUES (?, ?, ?, ?, true)
		`, id, file, time.Now(), checksumHex)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		log.Printf("Applied migration %s", file)
	}

	return nil
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
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO platforms (id, name, buying_power)
		VALUES (?, ?, ?);
	`, platform.ID, platform.Name, platform.BuyingPower)
	if err != nil {
		return err
	}

	// Fetch all models
	rows, err := tx.Query(`SELECT name FROM models;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var modelName string
		if err := rows.Scan(&modelName); err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO platform_models (platform_id, model_name, tokens) VALUES (?, ?, 0);`,
			platform.ID, modelName)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemovePlatform remove a platform by id
func RemovePlatform(id string) error {
	_, err := duckdbClient.Exec(`
	DELETE FROM platform_modelss WHERE platform_id = ?;
	DELETE FROM platforms WHERE id = ?;
	`, id, id)
	return err
}

// GetPlatformModelCost returns the cost of a model for a specific platform
func GetPlatformModelCost(platformID, modelName string) (int, error) {
	row := duckdbClient.QueryRow(`
        SELECT cost FROM platform_models
        WHERE platform_id = ? AND model_name = ?;
    `, platformID, modelName)

	var cost int
	err := row.Scan(&cost)
	return cost, err
}

// GetPlatform returns the buying power for a platform
func GetPlatform(platformID string) (Platform, error) {
	row := duckdbClient.QueryRow(`
        SELECT * FROM platforms
        WHERE id = ?;
    `, platformID)

	var id, name string
	var buying_power int

	err := row.Scan(&id, &name, &buying_power)
	if err != nil {
		return Platform{}, err
	}
	platform := Platform{
		ID:          id,
		Name:        name,
		BuyingPower: buying_power,
	}
	return platform, err
}

func ListPlatformModels() (models map[string][]ModelCost, err error) {
	models = make(map[string][]ModelCost)

	rows, err := duckdbClient.Query(`
	 SELECT pm.model_name, p.name AS platform_name, pm.tokens
	 FROM platform_models AS pm
	 JOIN platforms AS p
	 ON pm.platform_id = p.id;
	`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var platform_name, model_name string
		var tokens int

		err = rows.Scan(&model_name, &platform_name, &tokens)
		models[model_name] = append(models[model_name], ModelCost{
			PlatformName: platform_name,
			Tokens:       tokens,
		})
		if err != nil {
			break
		}

	}
	return
}

// SetPlatformModels set the platform model tokens
func SetPlatformModels(platformID, model string, tokens int) error {
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO platform_models (platform_id, model_name, tokens)
		VALUES (?, ?, ?) 
		ON CONFLICT DO UPDATE SET tokens = EXCLUDED.tokens;
	`, platformID, model, tokens)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// AddModel adds a model
func AddModel(model string) error {
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO models (name)
		VALUES (?);
	`, model)

	if err != nil {
		return err
	}
	// Fetch all platforms
	rows, err := tx.Query(`SELECT id FROM platforms;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var platformID string
		if err := rows.Scan(&platformID); err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO platform_models (platform_id, model_name, tokens) VALUES (?, ?, 0);`,
			platformID, model)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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
func RemoveModel(name string) error {
	_, err := duckdbClient.Exec(`
	DELETE FROM platform_models WHERE model_name = ?;
	DELETE FROM models WHERE name = ?;
	`, name, name)
	return err
}

// GetModel returns the model
func GetModel(name string) (string, error) {
	row := duckdbClient.QueryRow(`
        SELECT * FROM models
        WHERE name = ?;
    `, name)

	var mn string
	return name, row.Scan(&mn)
}

func AddHistory(hist History) error {
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO history (model_name, user_id, prompt)
		VALUES (?, ?, ?);
	`, hist.ModelName, hist.UserID, hist.Prompt)

	if err != nil {
		return err
	}
	return tx.Commit()
}

func ListHistory(index int) (history []History, err error) {
	rows, err := duckdbClient.Query(`SELECT * FROM history WHERE id > ? ORDER BY id ASC LIMIT 6;`, index)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var model_name, user_id, prompt string
		var id int
		err = rows.Scan(&id, &model_name, &prompt, &user_id)
		history = append(history, History{ID: id, ModelName: model_name, UserID: user_id, Prompt: prompt})

		if err != nil {
			break
		}
	}
	return

}

func CountHistory() int {
	row := duckdbClient.QueryRow(`SELECT currval('seq_history') AS currval;`)
	var index int32
	err := row.Scan(&index)
	if err != nil {
		slog.Error("Error fetching current sequence value:", slog.Any("err", err))
	}
	return int(index)
}

func GetHistory(id int) (history History, err error) {
	row := duckdbClient.QueryRow(`
        SELECT model_name, prompt FROM history
        WHERE id = ?;
    `, id)

	var model_name, prompt string
	err = row.Scan(&model_name, &prompt)

	return History{ID: id, ModelName: model_name, Prompt: prompt}, err
}

func GetContext(userID, modelName string) []int32 {
	row := duckdbClient.QueryRow(`
		SELECT context FROM contexts
		WHERE user_id = ? AND model_name = ?;
	`, userID, modelName)

	var raw []interface{}
	err := row.Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("Error fetching contexts:", slog.Any("err", err))
	}
	context := make([]int32, len(raw))
	for i, v := range raw {
		context[i] = v.(int32)
	}

	return context
}

func SetContext(userContext UserContext) error {
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO contexts (user_id, model_name, context, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?) 
		ON CONFLICT DO UPDATE SET 
		context = EXCLUDED.context,
		updated_at = EXCLUDED.updated_at;
	`, userContext.UserID, userContext.ModelName, userContext.Context, now, now)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// AddThread inserts a new thread record with an empty context slice.
func AddThread(modelName, systemPrompt, thread_id string) error {
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO threads (thread_id, model_name, system_prompt, context)
		VALUES (?, ?, ?, ?);
	`, thread_id, modelName, systemPrompt, []int{})

	if err != nil {
		return err
	}
	return tx.Commit()
}

func GetThread(id string) (Thread, error) {
	row := duckdbClient.QueryRow(`
		SELECT * FROM threads
		WHERE thread_id = ?;
	`, id)

	var thread_id, model_name, system_prompt string
	var context []interface{}
	err := row.Scan(&thread_id, &model_name, &system_prompt, &context)
	if err != nil {
		if err == sql.ErrNoRows {
			return Thread{}, err
		} else {
			slog.Error("Error fetching contexts:", slog.Any("err", err))
		}
	}

	contextSlice := make([]int32, len(context))
	for i, v := range context {
		contextSlice[i] = v.(int32)
	}

	return Thread{
		Context:   contextSlice,
		ThreadID:  thread_id,
		Prompt:    system_prompt,
		ModelName: model_name,
	}, nil
}

func UpdateThreadContext(id string, context []int32) error {
	tx, err := duckdbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE threads
		SET context = ?
		WHERE thread_id = ?;
		`, context, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
