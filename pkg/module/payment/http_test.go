package payment

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestCreatePaymentLink(t *testing.T) {
	randOrderId := rand.Intn(100000)
	CreateOrder(OrderDetail{
		OrderId:       fmt.Sprintf("TEST-%d", randOrderId),
		OrderAmount:   100,
		CustomerId:    "123",
		CustomerPhone: "+919639639879",
		ReturnUrl:     "http://192.168.1.4:8080/cart/payment/callback?order_id=66910e7311df548fa16d3cc2",
	})
}

// func TestExecuterazorpay(t *testing.T) {
// 	fmt.Println("hi")
// 	// id, err := executerazorpay()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	fmt.Println(id)

// }
