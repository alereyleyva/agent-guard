package gateway

import "github.com/google/uuid"

func generateTraceID() string {
	return uuid.New().String()
}
