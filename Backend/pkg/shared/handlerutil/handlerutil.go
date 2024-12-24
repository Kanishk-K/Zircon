package handlerutil

// HandlerRegister is an interface that defines a method to register routes
// Services should use RegisterRoutes to register their routes.
type HandlerRegister interface {
	RegisterRoutes()
}

// Takes comma separated list of HandlerRegister implementations and registers their routes
// Example:
// RegisterRoutes(authServiceHandler, userServiceHandler, jobQueueHandler, ...)
func RegisterRoutes(items ...HandlerRegister) {
	for _, item := range items {
		item.RegisterRoutes()
	}
}
