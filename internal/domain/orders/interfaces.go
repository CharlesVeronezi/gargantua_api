package orders

import (
	"github.com/discord-gophers/goapi-gen/runtime"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"net/http"
)

type ServerInterface interface {
	PostOrders(w http.ResponseWriter, r *http.Request) *Response
	GetOrders(w http.ResponseWriter, r *http.Request) *Response
	GetOrdersOrderID(w http.ResponseWriter, r *http.Request, orderID string) *Response
}

type ServerInterfaceWrapper struct {
	Handler          ServerInterface
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

type ServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

type ServerOption func(*ServerOptions)

func (siw *ServerInterfaceWrapper) PostOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := siw.Handler.PostOrders(w, r)
		if resp != nil {
			if resp.body != nil {
				render.Render(w, r, resp)
			} else {
				w.WriteHeader(resp.Code)
			}
		}
	})

	handler(w, r.WithContext(ctx))
}

func (siw *ServerInterfaceWrapper) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := siw.Handler.GetOrders(w, r)
		if resp != nil {
			if resp.body != nil {
				render.Render(w, r, resp)
			} else {
				w.WriteHeader(resp.Code)
			}
		}
	})

	handler(w, r.WithContext(ctx))
}

func (siw *ServerInterfaceWrapper) GetOrdersOrderID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// ------------- Path parameter "orderId" -------------
	var orderID string

	if err := runtime.BindStyledParameter("simple", false, "orderId", chi.URLParam(r, "orderId"), &orderID); err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{err, "orderId"})
		return
	}

	var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := siw.Handler.GetOrdersOrderID(w, r, orderID)
		if resp != nil {
			if resp.body != nil {
				render.Render(w, r, resp)
			} else {
				w.WriteHeader(resp.Code)
			}
		}
	})

	handler(w, r.WithContext(ctx))
}
