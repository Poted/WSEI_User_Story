# Implementacja User Stories dla sekcji Katalog Produktów

Ten projekt to przykładowa implementacja w języku Go dla zestawu User Stories dotyczących modułu "Katalog Produktów" w aplikacji do tworzenia list zakupów.

Aplikacja została zrealizowana jako proste API webowe, które udostępnia logikę biznesową poprzez endpointy HTTP.

## Omawiane User Stories

- Jako użytkownik, chcę wybierać jednostki (np. sztuki, gramy, litry) przy dodawaniu produktu, aby precyzyjnie określić ilość.
- Jako użytkownik, chcę mieć dostęp do katalogu najczęściej używanych produktów, aby szybciej dodawać je do listy zakupów.
- Jako użytkownik, chcę filtrować produkty według kategorii (np. nabiał, warzywa), aby łatwiej zarządzać zapasami.


## Zaimplementowane Funkcjonalności

W kodzie zaimplementowano następujące elementy:

* Proste API webowe (backend) oparte o standardowe pakiety Go (`net/http`).
* Endpointy do listowania wszystkich produktów oraz filtrowania ich po kategorii.
* Endpoint do dodawania produktów do listy zakupów.
* Logika zliczania popularności produktów (`UsageCount`).
* Endpoint zwracający najczęściej używane produkty.
* Zdefiniowany własny typ dla jednostek (`Unit`), działający jak "enum", dla bezpieczeństwa typów.
* Pokrycie kodu testami jednostkowymi dla trzech warstw: repozytorium, serwisu oraz handlerów API.

## Uruchomienie Aplikacji

Do uruchomienia aplikacji wymagana jest instalacja Go co na Windowsie można zrobić komendą:
```shell
winget install Go.Go
```
Następnie aplikację uruchamiamy komendą:
```shell
go run main.go
```

Serwer zostanie uruchomiony na adresie `http://localhost:8080`.

## Uruchomienie Testów

Uruchomienie testów jednostkowych, sprawdzających działanie aplikacji, wykonujemy komendą:

```bash
go test -v
```

Flaga `-v` (verbose) wyświetli szczegółowe wyniki dla każdego przypadku testowego.

## Dostępne Endpointy API

* `GET /` - Strona główna. Zwraca listę dostępnych endpointów jako placeholder.
* `GET /products` - Zwraca listę wszystkich produktów.
* `GET /products?category={nazwa}` - Zwraca produkty z podanej kategorii (np. `warzywa`).
* `GET /products/most-used` - Zwraca najczęściej używane produkty.
* `POST /shopping-list/add` - Dodaje produkt do listy. Wymaga `body` w formacie JSON, np. `{"ProductID": 1, "Unit": "l"}`.