package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMockProductRepository_FindByID(t *testing.T) {
	repo := NewMockProductRepository()

	testCases := []struct {
		name         string
		productID    int
		expectFound  bool
		expectedName string
	}{
		{"Znajdowanie istniejącego produktu", 1, true, "Mleko"},
		{"Próba znalezienia nieistniejącego produktu", 999, false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Szukanie produktu o ID: %d, oczekiwany wynik: %t", tc.productID, tc.expectFound)

			product, err := repo.FindByID(tc.productID)

			if tc.expectFound {
				if err != nil {
					t.Fatalf("oczekiwano znalezienia produktu, ale wystąpił błąd: %v", err)
				}
				if product.Name != tc.expectedName {
					t.Errorf("oczekiwano nazwy '%s', a otrzymano '%s'", tc.expectedName, product.Name)
				}
			} else {
				if err == nil {
					t.Fatal("oczekiwano błędu (produkt nie znaleziony), ale błędu nie było")
				}
			}
		})
	}
}

func TestShoppingService_AddProductToList(t *testing.T) {
	service := setupTestServer().shoppingService

	t.Run("Dodawanie poprawnego produktu", func(t *testing.T) {
		t.Log("Testowanie dodawania produktu z poprawną, dozwoloną jednostką...")
		initialProduct, _ := service.repo.FindByID(1)
		initialUsageCount := initialProduct.UsageCount

		_, err := service.AddProductToList(1, Liter)

		if err != nil {
			t.Fatalf("Dodawanie poprawnego produktu nie powinno zwrócić błędu, a zwróciło: %v", err)
		}
		if len(service.shoppingList) != 1 {
			t.Errorf("Oczekiwano 1 elementu na liście zakupów, a jest %d", len(service.shoppingList))
		}
		updatedProduct, _ := service.repo.FindByID(1)
		if updatedProduct.UsageCount != initialUsageCount+1 {
			t.Errorf("Licznik użycia produktu powinien zostać zwiększony o 1")
		}
	})

	t.Run("Dodawanie produktu z błędną jednostką", func(t *testing.T) {
		t.Log("Testowanie dodawania produktu z niedozwoloną jednostką...")
		_, err := service.AddProductToList(1, Kilogram)
		if err == nil {
			t.Fatal("Oczekiwano błędu przy dodawaniu produktu z nieprawidłową jednostką, ale go nie było")
		}
	})
}

func TestHandleGetProducts(t *testing.T) {
	server := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()

	t.Log("Testowanie endpointu GET /products...")
	server.handleGetProducts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("oczekiwano statusu 200 OK, a otrzymano %d", rr.Code)
	}

	var products []Product
	if err := json.Unmarshal(rr.Body.Bytes(), &products); err != nil {
		t.Fatalf("nie udało się zdekodować odpowiedzi JSON: %v", err)
	}

	if len(products) != 6 {
		t.Errorf("oczekiwano 6 produktów, a otrzymano %d", len(products))
	}
	t.Log("Endpoint GET /products działa poprawnie.")
}

func TestHandleAddItem(t *testing.T) {
	server := setupTestServer()
	requestBody := strings.NewReader(`{"ProductID": 1, "Unit": "l"}`)
	req := httptest.NewRequest(http.MethodPost, "/shopping-list/add", requestBody)
	rr := httptest.NewRecorder()

	t.Log("Testowanie endpointu POST /shopping-list/add...")
	server.handleAddItem(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("oczekiwano statusu 201 Created, a otrzymano %d", rr.Code)
	}

	var responseBody map[string]string
	json.Unmarshal(rr.Body.Bytes(), &responseBody)
	expectedMessage := "Dodano 'Mleko' do listy zakupów"
	if responseBody["message"] != expectedMessage {
		t.Errorf("oczekiwano wiadomości '%s', a otrzymano '%s'", expectedMessage, responseBody["message"])
	}
	t.Log("Endpoint POST /shopping-list/add działa poprawnie.")
}

func setupTestServer() *Server {
	repo := NewMockProductRepository()
	service := &ShoppingService{
		repo:         repo,
		shoppingList: []ShoppingItem{},
	}
	return &Server{
		shoppingService: service,
	}
}
