package httpapi

import (
	"context"
	"encoding/json"
	"gargantua/internal/domain/orders"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/rueidis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type API struct {
	client    *mongo.Client
	logger    *zap.Logger
	db        *mongo.Database
	validator *validator.Validate
	cache     rueidis.Client
}

func NewAPI(client *mongo.Client, logger *zap.Logger, redis rueidis.Client) *API {
	valid := validator.New()
	api := &API{
		client:    client,
		logger:    logger,
		db:        client.Database("gargantua"),
		validator: valid,
		cache:     redis,
	}
	return api
}

func (api *API) PostOrders(w http.ResponseWriter, r *http.Request) *orders.Response {
	var body orders.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return orders.ErrorJSON400Response(orders.Error{Message: "invalid JSON"})
	}

	if err := api.validator.Struct(body); err != nil {
		return orders.ErrorJSON400Response(orders.Error{Message: "invalid input: " + err.Error()})
	}

	order := &orders.CreateOrderRequest{
		UserID:          primitive.NewObjectID(),
		Products:        body.Products,
		TotalAmount:     body.TotalAmount,
		OrderStatus:     body.OrderStatus,
		PaymentMethod:   body.PaymentMethod,
		ShippingAddress: body.ShippingAddress,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result, err := api.db.Collection("orders").InsertOne(r.Context(), order)
	if err != nil {
		api.logger.Error("Failed to create order", zap.Error(err))
		return orders.ErrorJSON400Response(orders.Error{Message: "Failed to create order, try again"})
	}

	go func() {
		err := api.SaveOrderInCache(order)
		if err != nil {
			println("Error adding cache")
		}
	}()

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()
	response := orders.CreateOrderResponse{ID: insertedID}

	return orders.CreateOrderJSON201Response(response)

}

func (api *API) GetOrders(w http.ResponseWriter, r *http.Request) *orders.Response {

	cursor, err := api.db.Collection("orders").Find(r.Context(), bson.M{})
	if err != nil {
		api.logger.Error("Failed to fetch orders", zap.Error(err))
		return orders.ErrorJSON400Response(orders.Error{Message: "Failed to fetch orders"})
	}
	defer cursor.Close(r.Context())

	var ordersObj []orders.CreateOrderRequest

	for cursor.Next(r.Context()) {
		var order orders.CreateOrderRequest
		if err := cursor.Decode(&order); err != nil {
			api.logger.Error("Error decoding order", zap.Error(err))
			return orders.ErrorJSON400Response(orders.Error{Message: "Error decoding orders"})
		}
		ordersObj = append(ordersObj, order)
	}

	if err := cursor.Err(); err != nil {
		api.logger.Error("Error iterating over cursor", zap.Error(err))
		return orders.ErrorJSON400Response(orders.Error{Message: "Error fetching orders"})
	}

	return orders.OrdersJSON200Response(orders.GetOrdersResponse{Orders: ordersObj})
}

func (api *API) GetOrdersOrderID(w http.ResponseWriter, r *http.Request, orderID string) *orders.Response {

	orderFromCache, _ := api.GetOrderByIDInCache(orderID)
	println("orderFromCache:", orderFromCache)
	if orderFromCache != nil {
		println("Retorno do cache")
		return orders.UniqueOrderJSON200Response(*orderFromCache)
	}

	id, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		api.logger.Error("Invalid order ID format", zap.Error(err))
		return orders.ErrorJSON400Response(orders.Error{Message: "Invalid order ID format"})
	}

	var order orders.CreateOrderRequest
	err = api.db.Collection("orders").FindOne(r.Context(), bson.M{"_id": id}).Decode(&order)
	if err != nil {
		api.logger.Error("Failed to fetch order", zap.Error(err))
		return orders.ErrorJSON400Response(orders.Error{Message: "Failed to fetch order"})
	}
	println("Retorno do banco")

	return orders.UniqueOrderJSON200Response(order)

}

func (api *API) SaveOrderInCache(order *orders.CreateOrderRequest) error {
	ctx := context.Background()
	orderString, err := json.Marshal(order)
	if err != nil {
		return err
	}

	orderInCached := api.cache.B().Set().Key("order-id-" + order.ID.Hex()).Value(string(orderString)).Build()
	for _, resp := range api.cache.DoMulti(ctx, orderInCached) {
		if err := resp.Error(); err != nil {
			return err
		}
	}
	println("Cache adicionado:", orderString)
	return nil
}

func (api *API) GetOrderByIDInCache(id string) (*orders.CreateOrderRequest, error) {
	println("entrou na funcao get id cache")
	ctx := context.Background()
	orderResult, err := api.cache.Do(ctx, api.cache.B().Get().Key("order-id-"+id).Build()).ToString()
	if err != nil {
		return nil, err
	}
	var orderIdCache *orders.CreateOrderRequest
	err = json.Unmarshal([]byte(orderResult), &orderIdCache)
	if err != nil {
		return nil, err
	}
	println("order da funcao get id cache:", orderIdCache)
	println("memoria:", &orderIdCache)
	return orderIdCache, nil
}