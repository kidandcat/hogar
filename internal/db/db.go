package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kidandcat/hogar/internal/models"
	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB connection.
type DB struct {
	conn *sql.DB
}

// New opens the SQLite database and creates tables.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	d := &DB{conn: conn}
	if err := d.createTables(); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}
	return d, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) createTables() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS menus (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			week_start TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS menu_days (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
			day_name TEXT NOT NULL,
			lunch TEXT NOT NULL DEFAULT '',
			lunch_prep_time TEXT NOT NULL DEFAULT '',
			lunch_recipe_id INTEGER REFERENCES recipes(id) ON DELETE SET NULL,
			dinner TEXT NOT NULL DEFAULT '',
			dinner_prep_time TEXT NOT NULL DEFAULT '',
			dinner_recipe_id INTEGER REFERENCES recipes(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			ingredients TEXT NOT NULL DEFAULT '[]',
			steps TEXT NOT NULL DEFAULT '[]',
			prep_time TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_lists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			list_id INTEGER NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT '',
			checked INTEGER NOT NULL DEFAULT 0
		)`,
	}
	for _, s := range stmts {
		if _, err := d.conn.Exec(s); err != nil {
			return fmt.Errorf("exec %q: %w", s[:40], err)
		}
	}
	return nil
}

// ---------- Users ----------

// GetUserByUsername looks up a user by username.
func (d *DB) GetUserByUsername(username string) (*models.User, error) {
	u := &models.User{}
	err := d.conn.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUser inserts a new user.
func (d *DB) CreateUser(username, hashedPassword string) error {
	_, err := d.conn.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
	return err
}

// ---------- Menus ----------

// CreateMenu inserts a menu and its days.
func (d *DB) CreateMenu(m *models.Menu) (int64, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.DateTime)
	res, err := tx.Exec("INSERT INTO menus (week_start, created_at) VALUES (?, ?)", m.WeekStart, now)
	if err != nil {
		return 0, err
	}
	menuID, _ := res.LastInsertId()

	for _, day := range m.Days {
		_, err := tx.Exec(
			`INSERT INTO menu_days (menu_id, day_name, lunch, lunch_prep_time, lunch_recipe_id, dinner, dinner_prep_time, dinner_recipe_id)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			menuID, day.DayName, day.Lunch, day.LunchPrepTime, day.LunchRecipeID, day.Dinner, day.DinnerPrepTime, day.DinnerRecipeID,
		)
		if err != nil {
			return 0, err
		}
	}
	return menuID, tx.Commit()
}

// ListMenus returns all menus (without days).
func (d *DB) ListMenus() ([]models.Menu, error) {
	rows, err := d.conn.Query("SELECT id, week_start, created_at FROM menus ORDER BY week_start DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		if err := rows.Scan(&m.ID, &m.WeekStart, &m.CreatedAt); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	if menus == nil {
		menus = []models.Menu{}
	}
	return menus, rows.Err()
}

// GetMenu returns a menu with its days.
func (d *DB) GetMenu(id int64) (*models.Menu, error) {
	m := &models.Menu{}
	err := d.conn.QueryRow("SELECT id, week_start, created_at FROM menus WHERE id = ?", id).
		Scan(&m.ID, &m.WeekStart, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	days, err := d.getMenuDays(id)
	if err != nil {
		return nil, err
	}
	m.Days = days
	return m, nil
}

// GetCurrentMenu returns the menu for the current week (based on Monday).
func (d *DB) GetCurrentMenu() (*models.Menu, error) {
	now := time.Now()
	weekday := now.Weekday()
	// Go's Weekday: Sunday=0 ... Saturday=6
	// We want Monday as start. Offset: Monday=0, Tuesday=1, ..., Sunday=6
	offset := int(weekday) - 1
	if offset < 0 {
		offset = 6 // Sunday
	}
	monday := now.AddDate(0, 0, -offset)
	weekStart := monday.Format("2006-01-02")

	m := &models.Menu{}
	err := d.conn.QueryRow("SELECT id, week_start, created_at FROM menus WHERE week_start = ?", weekStart).
		Scan(&m.ID, &m.WeekStart, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	days, err := d.getMenuDays(m.ID)
	if err != nil {
		return nil, err
	}
	m.Days = days
	return m, nil
}

func (d *DB) getMenuDays(menuID int64) ([]models.MenuDay, error) {
	rows, err := d.conn.Query(
		`SELECT id, menu_id, day_name, lunch, lunch_prep_time, lunch_recipe_id, dinner, dinner_prep_time, dinner_recipe_id
		 FROM menu_days WHERE menu_id = ? ORDER BY id`, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []models.MenuDay
	for rows.Next() {
		var day models.MenuDay
		if err := rows.Scan(&day.ID, &day.MenuID, &day.DayName, &day.Lunch, &day.LunchPrepTime, &day.LunchRecipeID,
			&day.Dinner, &day.DinnerPrepTime, &day.DinnerRecipeID); err != nil {
			return nil, err
		}
		days = append(days, day)
	}
	if days == nil {
		days = []models.MenuDay{}
	}
	return days, rows.Err()
}

// DeleteMenu removes a menu and its days (cascade).
func (d *DB) DeleteMenu(id int64) error {
	_, err := d.conn.Exec("DELETE FROM menus WHERE id = ?", id)
	return err
}

// MenuCount returns the total number of menus.
func (d *DB) MenuCount() (int, error) {
	var count int
	err := d.conn.QueryRow("SELECT COUNT(*) FROM menus").Scan(&count)
	return count, err
}

// ---------- Recipes ----------

// CreateRecipe inserts a recipe, storing ingredients and steps as JSON.
func (d *DB) CreateRecipe(r *models.Recipe) (int64, error) {
	ingredientsJSON, _ := json.Marshal(r.Ingredients)
	stepsJSON, _ := json.Marshal(r.Steps)
	now := time.Now().UTC().Format(time.DateTime)

	res, err := d.conn.Exec(
		"INSERT INTO recipes (name, ingredients, steps, prep_time, created_at) VALUES (?, ?, ?, ?, ?)",
		r.Name, string(ingredientsJSON), string(stepsJSON), r.PrepTime, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListRecipes returns all recipes.
func (d *DB) ListRecipes() ([]models.Recipe, error) {
	rows, err := d.conn.Query("SELECT id, name, ingredients, steps, prep_time, created_at FROM recipes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		r, err := scanRecipe(rows)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *r)
	}
	if recipes == nil {
		recipes = []models.Recipe{}
	}
	return recipes, rows.Err()
}

// GetRecipe returns a single recipe by ID.
func (d *DB) GetRecipe(id int64) (*models.Recipe, error) {
	row := d.conn.QueryRow("SELECT id, name, ingredients, steps, prep_time, created_at FROM recipes WHERE id = ?", id)
	r, err := scanRecipeRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

// DeleteRecipe removes a recipe by ID.
func (d *DB) DeleteRecipe(id int64) error {
	_, err := d.conn.Exec("DELETE FROM recipes WHERE id = ?", id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRecipe(rows *sql.Rows) (*models.Recipe, error) {
	var r models.Recipe
	var ingredientsJSON, stepsJSON string
	if err := rows.Scan(&r.ID, &r.Name, &ingredientsJSON, &stepsJSON, &r.PrepTime, &r.CreatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(ingredientsJSON), &r.Ingredients)
	json.Unmarshal([]byte(stepsJSON), &r.Steps)
	if r.Ingredients == nil {
		r.Ingredients = []string{}
	}
	if r.Steps == nil {
		r.Steps = []string{}
	}
	return &r, nil
}

func scanRecipeRow(row *sql.Row) (*models.Recipe, error) {
	var r models.Recipe
	var ingredientsJSON, stepsJSON string
	if err := row.Scan(&r.ID, &r.Name, &ingredientsJSON, &stepsJSON, &r.PrepTime, &r.CreatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(ingredientsJSON), &r.Ingredients)
	json.Unmarshal([]byte(stepsJSON), &r.Steps)
	if r.Ingredients == nil {
		r.Ingredients = []string{}
	}
	if r.Steps == nil {
		r.Steps = []string{}
	}
	return &r, nil
}

// ---------- Shopping Lists ----------

// CreateShoppingList inserts a shopping list and its items.
func (d *DB) CreateShoppingList(sl *models.ShoppingList) (int64, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.DateTime)
	res, err := tx.Exec("INSERT INTO shopping_lists (name, created_at) VALUES (?, ?)", sl.Name, now)
	if err != nil {
		return 0, err
	}
	listID, _ := res.LastInsertId()

	for _, item := range sl.Items {
		_, err := tx.Exec(
			"INSERT INTO shopping_items (list_id, name, category, checked) VALUES (?, ?, ?, 0)",
			listID, item.Name, item.Category,
		)
		if err != nil {
			return 0, err
		}
	}
	return listID, tx.Commit()
}

// ListShoppingLists returns all shopping lists (without items).
func (d *DB) ListShoppingLists() ([]models.ShoppingList, error) {
	rows, err := d.conn.Query("SELECT id, name, created_at FROM shopping_lists ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.ShoppingList
	for rows.Next() {
		var sl models.ShoppingList
		if err := rows.Scan(&sl.ID, &sl.Name, &sl.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, sl)
	}
	if lists == nil {
		lists = []models.ShoppingList{}
	}
	return lists, rows.Err()
}

// GetShoppingList returns a shopping list with its items.
func (d *DB) GetShoppingList(id int64) (*models.ShoppingList, error) {
	sl := &models.ShoppingList{}
	err := d.conn.QueryRow("SELECT id, name, created_at FROM shopping_lists WHERE id = ?", id).
		Scan(&sl.ID, &sl.Name, &sl.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := d.conn.Query(
		"SELECT id, list_id, name, category, checked FROM shopping_items WHERE list_id = ? ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.ShoppingItem
		var checked int
		if err := rows.Scan(&item.ID, &item.ListID, &item.Name, &item.Category, &checked); err != nil {
			return nil, err
		}
		item.Checked = checked != 0
		sl.Items = append(sl.Items, item)
	}
	if sl.Items == nil {
		sl.Items = []models.ShoppingItem{}
	}
	return sl, rows.Err()
}

// ToggleShoppingItem updates the checked state of a shopping item.
func (d *DB) ToggleShoppingItem(listID, itemID int64, checked bool) error {
	val := 0
	if checked {
		val = 1
	}
	res, err := d.conn.Exec("UPDATE shopping_items SET checked = ? WHERE id = ? AND list_id = ?", val, itemID, listID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteShoppingList removes a shopping list and its items (cascade).
func (d *DB) DeleteShoppingList(id int64) error {
	_, err := d.conn.Exec("DELETE FROM shopping_lists WHERE id = ?", id)
	return err
}
