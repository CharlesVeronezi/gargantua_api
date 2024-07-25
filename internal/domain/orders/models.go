package orders

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type OrderID primitive.ObjectID

type Product struct {
	ProductID primitive.ObjectID `bson:"product_id,omitempty"`
	Quantity  int                `bson:"quantity"`
	Price     float64            `bson:"price"`
}

type Address struct {
	Street  string `bson:"street"`
	City    string `bson:"city"`
	State   string `bson:"state"`
	Zip     string `bson:"zip"`
	Country string `bson:"country"`
}

type CreateOrderRequest struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          primitive.ObjectID `bson:"userId,omitempty"`
	Products        []Product          `bson:"products,omitempty"`
	TotalAmount     float64            `bson:"total_amount"`
	OrderStatus     string             `bson:"order_status"`
	PaymentMethod   string             `bson:"payment_method"`
	ShippingAddress Address            `bson:"shipping_address"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}
