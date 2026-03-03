// Package contextkeys определяет ключи для значений в gin.Context и context.Context.
package contextkeys

// UserIDKey — ключ в gin.Context для user_id (Gin принимает только string).
const UserIDKey = "user_id"

// contextKey — тип ключа для context.Context, чтобы избежать коллизий (go vet).
type contextKey string

// UserIDContextKey — ключ для user_id в context.Context (gRPC).
const UserIDContextKey contextKey = "user_id"
