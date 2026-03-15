package handlers

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/kidandcat/hogar/internal/auth"
	"github.com/kidandcat/hogar/internal/db"
	"github.com/kidandcat/hogar/internal/models"
)

// View models for templates

type ShoppingCategory struct {
	Name  string
	Items []models.ShoppingItem
}

type ShoppingActiveList struct {
	ID         int64
	Name       string
	Categories []ShoppingCategory
}

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	DB   *db.DB
	Tmpl *template.Template
}

// New creates a Handler with parsed templates from the given filesystem.
func New(database *db.DB, tmplFS fs.FS) *Handler {
	tmpl := template.Must(template.ParseFS(tmplFS, "templates/*.html"))
	return &Handler{DB: database, Tmpl: tmpl}
}

// RegisterRoutes sets up all routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Page routes
	mux.HandleFunc("GET /login", h.loginPage)
	mux.HandleFunc("POST /login", h.loginSubmit)
	mux.HandleFunc("GET /logout", h.logout)
	mux.HandleFunc("GET /", h.dashboard)
	mux.HandleFunc("GET /menu", h.menuPage)
	mux.HandleFunc("GET /shopping", h.shoppingPage)
	mux.HandleFunc("GET /recipes", h.recipesPage)

	// API routes - Menus
	mux.HandleFunc("POST /api/menus", h.createMenu)
	mux.HandleFunc("GET /api/menus", h.listMenus)
	mux.HandleFunc("GET /api/menus/current", h.currentMenu)
	mux.HandleFunc("GET /api/menus/{id}", h.getMenu)
	mux.HandleFunc("DELETE /api/menus/{id}", h.deleteMenu)

	// API routes - Shopping Lists
	mux.HandleFunc("GET /api/shopping-lists", h.listShoppingLists)
	mux.HandleFunc("GET /api/shopping-lists/{id}", h.getShoppingList)
	mux.HandleFunc("POST /api/shopping-lists", h.createShoppingList)
	mux.HandleFunc("PATCH /api/shopping-lists/{id}/items/{itemId}", h.toggleShoppingItem)
	mux.HandleFunc("DELETE /api/shopping-lists/{id}", h.deleteShoppingList)

	// API routes - Recipes
	mux.HandleFunc("POST /api/recipes", h.createRecipe)
	mux.HandleFunc("GET /api/recipes", h.listRecipes)
	mux.HandleFunc("GET /api/recipes/{id}", h.getRecipe)
	mux.HandleFunc("DELETE /api/recipes/{id}", h.deleteRecipe)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pathID(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(r.PathValue(name), 10, 64)
}

// --- Page Handlers ---

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to dashboard
	if _, ok := auth.ValidateRequest(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "login.html", map[string]any{})
}

func (h *Handler) loginSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if auth.Login(h.DB, w, username, password) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Invalid credentials"})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	auth.Logout(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	menu, _ := h.DB.GetCurrentMenu()
	h.Tmpl.ExecuteTemplate(w, "dashboard.html", map[string]any{"Menu": menu})
}

func (h *Handler) menuPage(w http.ResponseWriter, r *http.Request) {
	menus, _ := h.DB.ListMenus()
	// Load days for each menu so the template can display them
	for i := range menus {
		full, _ := h.DB.GetMenu(menus[i].ID)
		if full != nil {
			menus[i].Days = full.Days
		}
	}
	h.Tmpl.ExecuteTemplate(w, "menu.html", map[string]any{"Menus": menus})
}

func (h *Handler) shoppingPage(w http.ResponseWriter, r *http.Request) {
	lists, _ := h.DB.ListShoppingLists()
	data := map[string]any{"Lists": lists}

	if len(lists) > 0 {
		// Select active list from query param, default to first
		selectedID := lists[0].ID
		if idStr := r.URL.Query().Get("list"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				selectedID = id
			}
		}

		sl, _ := h.DB.GetShoppingList(selectedID)
		if sl != nil {
			categoryOrder := []string{"Carnes", "Lácteos", "Frutas/Verduras", "Congelados", "Despensa", "Limpieza", "Otros"}
			categoryMap := make(map[string][]models.ShoppingItem)
			for _, item := range sl.Items {
				cat := item.Category
				if cat == "" {
					cat = "Otros"
				}
				categoryMap[cat] = append(categoryMap[cat], item)
			}

			var categories []ShoppingCategory
			for _, name := range categoryOrder {
				if items, ok := categoryMap[name]; ok {
					categories = append(categories, ShoppingCategory{Name: name, Items: items})
				}
			}
			// Add any extra categories not in the predefined order
			for cat, items := range categoryMap {
				found := false
				for _, name := range categoryOrder {
					if cat == name {
						found = true
						break
					}
				}
				if !found {
					categories = append(categories, ShoppingCategory{Name: cat, Items: items})
				}
			}

			data["ActiveList"] = &ShoppingActiveList{
				ID:         sl.ID,
				Name:       sl.Name,
				Categories: categories,
			}
		}
	}

	h.Tmpl.ExecuteTemplate(w, "shopping.html", data)
}

func (h *Handler) recipesPage(w http.ResponseWriter, r *http.Request) {
	recipes, _ := h.DB.ListRecipes()
	h.Tmpl.ExecuteTemplate(w, "recipes.html", map[string]any{"Recipes": recipes})
}

// --- API: Menus ---

func (h *Handler) createMenu(w http.ResponseWriter, r *http.Request) {
	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if menu.WeekStart == "" {
		writeError(w, http.StatusBadRequest, "week_start is required")
		return
	}

	id, err := h.DB.CreateMenu(&menu)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	created, _ := h.DB.GetMenu(id)
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listMenus(w http.ResponseWriter, r *http.Request) {
	menus, err := h.DB.ListMenus()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, menus)
}

func (h *Handler) currentMenu(w http.ResponseWriter, r *http.Request) {
	menu, err := h.DB.GetCurrentMenu()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if menu == nil {
		writeError(w, http.StatusNotFound, "no menu for current week")
		return
	}
	writeJSON(w, http.StatusOK, menu)
}

func (h *Handler) getMenu(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	menu, err := h.DB.GetMenu(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if menu == nil {
		writeError(w, http.StatusNotFound, "menu not found")
		return
	}
	writeJSON(w, http.StatusOK, menu)
}

func (h *Handler) deleteMenu(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.DB.DeleteMenu(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- API: Shopping Lists ---

func (h *Handler) listShoppingLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.DB.ListShoppingLists()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, lists)
}

func (h *Handler) getShoppingList(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	sl, err := h.DB.GetShoppingList(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sl == nil {
		writeError(w, http.StatusNotFound, "shopping list not found")
		return
	}
	writeJSON(w, http.StatusOK, sl)
}

func (h *Handler) createShoppingList(w http.ResponseWriter, r *http.Request) {
	var sl models.ShoppingList
	if err := json.NewDecoder(r.Body).Decode(&sl); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if sl.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	id, err := h.DB.CreateShoppingList(&sl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	created, _ := h.DB.GetShoppingList(id)
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) toggleShoppingItem(w http.ResponseWriter, r *http.Request) {
	listID, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}
	itemID, err := pathID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var body struct {
		Checked bool `json:"checked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.DB.ToggleShoppingItem(listID, itemID, body.Checked); err != nil {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"checked": body.Checked})
}

func (h *Handler) deleteShoppingList(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.DB.DeleteShoppingList(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- API: Recipes ---

func (h *Handler) createRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if recipe.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if recipe.Ingredients == nil {
		recipe.Ingredients = []string{}
	}
	if recipe.Steps == nil {
		recipe.Steps = []string{}
	}

	id, err := h.DB.CreateRecipe(&recipe)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	created, _ := h.DB.GetRecipe(id)
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.DB.ListRecipes()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, recipes)
}

func (h *Handler) getRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	recipe, err := h.DB.GetRecipe(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if recipe == nil {
		writeError(w, http.StatusNotFound, "recipe not found")
		return
	}
	writeJSON(w, http.StatusOK, recipe)
}

func (h *Handler) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.DB.DeleteRecipe(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
