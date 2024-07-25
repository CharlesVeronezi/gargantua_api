package httpapi

import (
	"gargantua/internal/domain/orders"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Handler(si orders.ServerInterface, opts ...orders.ServerOption) http.Handler {
	options := &orders.ServerOptions{
		BaseURL:    "/",
		BaseRouter: chi.NewRouter(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	}

	for _, f := range opts {
		f(options)
	}

	r := options.BaseRouter
	wrapper := orders.ServerInterfaceWrapper{
		Handler:          si,
		ErrorHandlerFunc: options.ErrorHandlerFunc,
	}

	r.Route(options.BaseURL, func(r chi.Router) {
		r.Post("/orders", wrapper.PostOrders)
		r.Get("/orders", wrapper.GetOrders)
		r.Get("/orders/{orderId}", wrapper.GetOrdersOrderID)
	})
	return r
}
