package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	cafeCount := func(city string) int {
		ln := len(cafeList[city])
		if ln < 100 {
			return ln
		}
		return 100
	}

	requests := []struct {
		city  string
		count int
		want  int
	}{
		{"moscow", 0, 0},
		{"moscow", 1, 1},
		{"moscow", 2, 2},
		{"moscow", 100, cafeCount("moscow")},

		{"tula", 0, 0},
		{"tula", 1, 1},
		{"tula", 2, 2},
		{"tula", 100, cafeCount("tula")},
	}

	for _, test := range requests {
		handler := http.HandlerFunc(mainHandle)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/cafe?city=%s&count=%d", test.city, test.count), nil)
		handler.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)

		coffeeCount := func() int {
			cafeLst := strings.Split(strings.TrimSpace(resp.Body.String()), ",")
			if cafeLst[0] == "" {
				return 0
			}
			return len(cafeLst)
		}()
		assert.Equal(t, test.want, coffeeCount)
	}

}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, test := range requests {
		handler := http.HandlerFunc(mainHandle)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet, fmt.Sprintf("/cafe?city=moscow&search=%s", test.search), nil,
		)
		handler.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)

		cafeList := func() []string {
			cafeList := strings.Split(strings.TrimSpace(resp.Body.String()), ",")
			if cafeList[0] == "" {
				return []string{}
			}
			return cafeList
		}()

		var actualCount int
		for _, cafe := range cafeList {
			if strings.Contains(strings.ToLower(cafe), test.search) {
				actualCount++
			}
		}
		assert.Equal(t, test.wantCount, actualCount)
	}
}
