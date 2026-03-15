package models

// User represents an authenticated user.
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // bcrypt hashed, never exposed in JSON
}

// Menu represents a weekly meal plan.
type Menu struct {
	ID        int64     `json:"id"`
	WeekStart string    `json:"week_start"` // ISO date of Monday (YYYY-MM-DD)
	Days      []MenuDay `json:"days"`
	CreatedAt string    `json:"created_at"`
}

// MenuDay represents meals for a single day within a menu.
type MenuDay struct {
	ID             int64  `json:"id"`
	MenuID         int64  `json:"menu_id"`
	DayName        string `json:"day_name"` // Lunes, Martes, etc.
	Lunch          string `json:"lunch"`
	LunchPrepTime  string `json:"lunch_prep_time"`
	LunchRecipeID  *int64 `json:"lunch_recipe_id,omitempty"`
	Dinner         string `json:"dinner"`
	DinnerPrepTime string `json:"dinner_prep_time"`
	DinnerRecipeID *int64 `json:"dinner_recipe_id,omitempty"`
}

// Recipe represents a cooking recipe.
type Recipe struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
	PrepTime    string   `json:"prep_time"`
	CreatedAt   string   `json:"created_at"`
}

// ShoppingList represents a list of items to buy.
type ShoppingList struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Items     []ShoppingItem `json:"items"`
	CreatedAt string         `json:"created_at"`
}

// ShoppingItem represents a single item in a shopping list.
type ShoppingItem struct {
	ID       int64  `json:"id"`
	ListID   int64  `json:"list_id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Checked  bool   `json:"checked"`
}
