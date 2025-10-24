package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Unit string

const (
	Piece     Unit = "szt."
	Kilogram  Unit = "kg"
	Gram      Unit = "g"
	Liter     Unit = "l"
	Mililiter Unit = "ml"
)

type Product struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	AvailableUnits []Unit `json:"availableUnits"`
	UsageCount     int    `json:"-"`
}

type ProductRepository interface {
	GetAll() ([]Product, error)
	GetByCategory(category string) ([]Product, error)
	FindByID(id int) (*Product, error)
	GetMostUsed(limit int) ([]Product, error)
}

type MockProductRepository struct {
	products []Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		products: []Product{
			{ID: 1, Name: "Mleko", Category: "nabiał", AvailableUnits: []Unit{Liter, Mililiter}},
			{ID: 2, Name: "Ser żółty", Category: "nabiał", AvailableUnits: []Unit{Kilogram, Gram}},
			{ID: 3, Name: "Jogurt naturalny", Category: "nabiał", AvailableUnits: []Unit{Piece, Gram}},
			{ID: 4, Name: "Pomidor", Category: "warzywa", AvailableUnits: []Unit{Piece, Kilogram}},
			{ID: 5, Name: "Ogórek", Category: "warzywa", AvailableUnits: []Unit{Piece, Kilogram}},
			{ID: 6, Name: "Chleb", Category: "pieczywo", AvailableUnits: []Unit{Piece}},
		},
	}
}

func (m *MockProductRepository) GetAll() ([]Product, error) {
	return m.products, nil
}

func (m *MockProductRepository) FindByID(id int) (*Product, error) {
	for i := range m.products {
		if m.products[i].ID == id {
			return &m.products[i], nil
		}
	}
	return nil, fmt.Errorf("produkt o ID %d nie został znaleziony", id)
}

func (m *MockProductRepository) GetByCategory(category string) ([]Product, error) {
	var filtered []Product
	for _, p := range m.products {
		if strings.EqualFold(p.Category, category) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (m *MockProductRepository) GetMostUsed(limit int) ([]Product, error) {

	productsCopy := make([]Product, len(m.products))
	copy(productsCopy, m.products)

	sort.Slice(productsCopy, func(i, j int) bool {
		return productsCopy[i].UsageCount > productsCopy[j].UsageCount
	})

	if limit > len(productsCopy) {
		limit = len(productsCopy)
	}

	return productsCopy[:limit], nil
}

type ShoppingItem struct {
	ProductID int
	Quantity  float64
	Unit      Unit
}

type ShoppingService struct {
	repo         ProductRepository
	shoppingList []ShoppingItem
}

func (s *ShoppingService) AddProductToList(productID int, unit Unit) (*Product, error) {

	product, err := s.repo.FindByID(productID)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(product.AvailableUnits, unit) {
		return nil, fmt.Errorf("produkt %s nie jest dostępny w jednostce %s", product.Name, unit)
	}

	s.shoppingList = append(s.shoppingList, ShoppingItem{
		ProductID: productID,
		Quantity:  1,
		Unit:      unit,
	})

	product.UsageCount++
	return product, nil
}

type Server struct {
	shoppingService *ShoppingService
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (s *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	var products []Product
	var err error

	if category != "" {
		products, err = s.shoppingService.repo.GetByCategory(category)
	} else {
		products, err = s.shoppingService.repo.GetAll()
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, products)
}

func (s *Server) handleGetMostUsed(w http.ResponseWriter, r *http.Request) {
	limit := 3
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	products, err := s.shoppingService.repo.GetMostUsed(limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, products)
}

func (s *Server) handleAddItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req := ShoppingItem{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "wrong json format")
		return
	}

	product, err := s.shoppingService.AddProductToList(req.ProductID, req.Unit)
	if err != nil {
		if strings.Contains(err.Error(), "nie został znaleziony") {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{
		"message": fmt.Sprintf("Dodano '%s' do listy zakupów", product.Name),
	})
}

func main() {

	productRepo := NewMockProductRepository()

	shoppingService := &ShoppingService{
		repo:         productRepo,
		shoppingList: []ShoppingItem{},
	}

	server := &Server{
		shoppingService: shoppingService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(
			[]byte("Przykładowe użycie dostępnych endpointów:\n" +
				"  GET /products\n" +
				"  GET /products?category=warzywa\n" +
				"  GET /products/most-used\n" +
				"  POST /shopping-list/add\n",
			),
		)
	})
	mux.HandleFunc("/products", server.handleGetProducts)
	mux.HandleFunc("/products/most-used", server.handleGetMostUsed)
	mux.HandleFunc("/shopping-list/add", server.handleAddItem)

	port := "8080"

	if err := http.ListenAndServe(":"+port, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Błąd serwera HTTP: %v", err)
	}
}
