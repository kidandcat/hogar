package db

import "github.com/kidandcat/hogar/internal/models"

// Seed populates the database with initial data if no menus exist.
func (d *DB) Seed() error {
	count, err := d.MenuCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	// Create recipes for each meal
	recipes := []models.Recipe{
		{
			Name:        "Pechuga plancha + quinoa + tomate + aguacate",
			Ingredients: []string{"Pechuga de pollo", "Quinoa", "Tomate", "Aguacate", "Aceite de oliva", "Sal"},
			Steps:       []string{"Cocinar la quinoa según instrucciones", "Hacer la pechuga a la plancha con sal", "Cortar tomate y aguacate", "Servir todo junto"},
			PrepTime:    "10min",
		},
		{
			Name:        "Tortilla francesa jamón york + palitos zanahoria",
			Ingredients: []string{"Huevos", "Jamón york", "Zanahoria", "Sal", "Aceite"},
			Steps:       []string{"Batir huevos con sal", "Hacer tortilla francesa con jamón york", "Pelar y cortar zanahoria en palitos", "Servir"},
			PrepTime:    "10min",
		},
		{
			Name:        "Salmón al horno + boniato micro + ensalada",
			Ingredients: []string{"Salmón fresco", "Boniato", "Lechuga", "Tomate", "Aceite de oliva", "Sal", "Limón"},
			Steps:       []string{"Precalentar horno a 200°C", "Hornear salmón 15 min con sal y limón", "Boniato al microondas 8 min", "Preparar ensalada"},
			PrepTime:    "20min",
		},
		{
			Name:        "Macarrones tomate + albóndigas Hacendado",
			Ingredients: []string{"Macarrones", "Salsa de tomate", "Albóndigas Hacendado", "Sal"},
			Steps:       []string{"Hervir macarrones según instrucciones", "Calentar albóndigas en sartén", "Calentar salsa de tomate", "Mezclar todo y servir"},
			PrepTime:    "15min",
		},
		{
			Name:        "Revuelto huevos con gambas al ajillo + tostada tomate",
			Ingredients: []string{"Huevos", "Gambas peladas", "Ajo", "Pan", "Tomate", "Aceite de oliva", "Sal"},
			Steps:       []string{"Hacer gambas al ajillo en sartén", "Añadir huevos batidos y revolver", "Tostar pan y untar tomate rallado", "Servir"},
			PrepTime:    "10min",
		},
		{
			Name:        "Crema verduras Hacendado + taquitos pollo plancha",
			Ingredients: []string{"Crema de verduras Hacendado", "Pechuga de pollo", "Sal", "Aceite de oliva"},
			Steps:       []string{"Calentar crema de verduras", "Cortar pollo en taquitos y hacer a la plancha", "Servir crema con taquitos encima"},
			PrepTime:    "10min",
		},
		{
			Name:        "Poke bowl exprés: arroz micro + salmón ahumado + aguacate + pepino + soja",
			Ingredients: []string{"Arroz para microondas", "Salmón ahumado", "Aguacate", "Pepino", "Salsa de soja", "Semillas de sésamo"},
			Steps:       []string{"Calentar arroz en microondas", "Cortar salmón, aguacate y pepino", "Montar bowl con arroz de base", "Añadir toppings y salsa de soja"},
			PrepTime:    "5min",
		},
		{
			Name:        "Salchichas pollo plancha + puré patata + guisantes",
			Ingredients: []string{"Salchichas de pollo", "Puré de patata instantáneo", "Guisantes congelados", "Mantequilla", "Leche", "Sal"},
			Steps:       []string{"Hacer salchichas a la plancha", "Preparar puré según instrucciones", "Hervir guisantes 5 min", "Servir todo junto"},
			PrepTime:    "15min",
		},
		{
			Name:        "Wrap pollo: tortilla mexicana + pollo plancha + lechuga + tomate + yogur griego",
			Ingredients: []string{"Tortillas mexicanas", "Pechuga de pollo", "Lechuga", "Tomate", "Yogur griego", "Sal"},
			Steps:       []string{"Hacer pollo a la plancha y cortar en tiras", "Cortar lechuga y tomate", "Calentar tortilla", "Montar wrap con todos los ingredientes y yogur"},
			PrepTime:    "10min",
		},
		{
			Name:        "Pizza fresca Hacendado + ensalada",
			Ingredients: []string{"Pizza fresca Hacendado", "Lechuga", "Tomate", "Aceite de oliva", "Sal"},
			Steps:       []string{"Precalentar horno según instrucciones de la pizza", "Hornear pizza", "Preparar ensalada", "Servir"},
			PrepTime:    "10min",
		},
		{
			Name:        "Huevos rotos con jamón serrano + patatas bolsa",
			Ingredients: []string{"Huevos", "Jamón serrano", "Patatas de bolsa para freír", "Aceite de oliva", "Sal"},
			Steps:       []string{"Freír patatas de bolsa", "Freír huevos", "Romper huevos sobre patatas", "Añadir jamón serrano por encima"},
			PrepTime:    "15min",
		},
		{
			Name:        "Hamburguesas caseras pan brioche + boniato horno",
			Ingredients: []string{"Carne picada de ternera", "Pan brioche", "Lechuga", "Tomate", "Queso", "Boniato", "Sal", "Aceite de oliva"},
			Steps:       []string{"Cortar boniato y hornear a 200°C 20 min", "Formar hamburguesas con la carne y sal", "Hacer a la plancha", "Montar en pan brioche con lechuga, tomate y queso"},
			PrepTime:    "25min",
		},
		{
			Name:        "Pollo entero marinado Hacendado al horno + patatas",
			Ingredients: []string{"Pollo entero marinado Hacendado", "Patatas", "Aceite de oliva", "Sal"},
			Steps:       []string{"Precalentar horno a 190°C", "Pelar y cortar patatas", "Colocar pollo y patatas en bandeja", "Hornear 1 hora", "Servir"},
			PrepTime:    "5min prep, 1h horno",
		},
		{
			Name:        "Huevos revueltos jamón york + pan tomate",
			Ingredients: []string{"Huevos", "Jamón york", "Pan", "Tomate", "Aceite de oliva", "Sal"},
			Steps:       []string{"Batir huevos con sal", "Hacer revuelto con taquitos de jamón york", "Tostar pan y untar tomate", "Servir"},
			PrepTime:    "5min",
		},
	}

	recipeIDs := make([]int64, len(recipes))
	for i, r := range recipes {
		id, err := d.CreateRecipe(&r)
		if err != nil {
			return err
		}
		recipeIDs[i] = id
	}

	// Create menu for week starting 2026-03-16
	menu := &models.Menu{
		WeekStart: "2026-03-16",
		Days: []models.MenuDay{
			{
				DayName:        "Lunes",
				Lunch:          "Pechuga plancha + quinoa + tomate + aguacate",
				LunchPrepTime:  "10min",
				LunchRecipeID:  &recipeIDs[0],
				Dinner:         "Tortilla francesa jamón york + palitos zanahoria",
				DinnerPrepTime: "10min",
				DinnerRecipeID: &recipeIDs[1],
			},
			{
				DayName:        "Martes",
				Lunch:          "Salmón al horno + boniato micro + ensalada",
				LunchPrepTime:  "20min",
				LunchRecipeID:  &recipeIDs[2],
				Dinner:         "Macarrones tomate + albóndigas Hacendado",
				DinnerPrepTime: "15min",
				DinnerRecipeID: &recipeIDs[3],
			},
			{
				DayName:        "Miércoles",
				Lunch:          "Revuelto huevos con gambas al ajillo + tostada tomate",
				LunchPrepTime:  "10min",
				LunchRecipeID:  &recipeIDs[4],
				Dinner:         "Crema verduras Hacendado + taquitos pollo plancha",
				DinnerPrepTime: "10min",
				DinnerRecipeID: &recipeIDs[5],
			},
			{
				DayName:        "Jueves",
				Lunch:          "Poke bowl exprés: arroz micro + salmón ahumado + aguacate + pepino + soja",
				LunchPrepTime:  "5min",
				LunchRecipeID:  &recipeIDs[6],
				Dinner:         "Salchichas pollo plancha + puré patata + guisantes",
				DinnerPrepTime: "15min",
				DinnerRecipeID: &recipeIDs[7],
			},
			{
				DayName:        "Viernes",
				Lunch:          "Wrap pollo: tortilla mexicana + pollo plancha + lechuga + tomate + yogur griego",
				LunchPrepTime:  "10min",
				LunchRecipeID:  &recipeIDs[8],
				Dinner:         "Pizza fresca Hacendado + ensalada",
				DinnerPrepTime: "10min",
				DinnerRecipeID: &recipeIDs[9],
			},
			{
				DayName:        "Sábado",
				Lunch:          "Huevos rotos con jamón serrano + patatas bolsa",
				LunchPrepTime:  "15min",
				LunchRecipeID:  &recipeIDs[10],
				Dinner:         "Hamburguesas caseras pan brioche + boniato horno",
				DinnerPrepTime: "25min",
				DinnerRecipeID: &recipeIDs[11],
			},
			{
				DayName:        "Domingo",
				Lunch:          "Pollo entero marinado Hacendado al horno + patatas",
				LunchPrepTime:  "5min prep, 1h horno",
				LunchRecipeID:  &recipeIDs[12],
				Dinner:         "Huevos revueltos jamón york + pan tomate",
				DinnerPrepTime: "5min",
				DinnerRecipeID: &recipeIDs[13],
			},
		},
	}

	_, err = d.CreateMenu(menu)
	return err
}
