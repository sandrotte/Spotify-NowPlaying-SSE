package main

import "net/http"

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		responseWriter.Header().Set("Vary", "Origin")

		if request.Method == http.MethodOptions {
			responseWriter.WriteHeader(http.StatusNoContent)

			return
		}

		next(responseWriter, request)
	}
}