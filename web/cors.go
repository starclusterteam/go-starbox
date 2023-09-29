package web

import (
	"github.com/rs/cors"
)

func CorsHandler(allowedOrigins []string, allowedHeaders []string) *cors.Cors {
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "HEAD", "OPTIONS"},
		AllowedHeaders:   append([]string{"Content-Type"}, allowedHeaders...),
		AllowCredentials: true,
	})

	return c
}

