package payment

import (
	"fmt"
	"log"
	"os"
	"src/l"
	"time"

	cashfree "github.com/cashfree/cashfree-pg/v4"
)

func CreatePGLink() (*cashfree.LinkEntity, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// 10 minute link expiry time
	clientId := os.Getenv("CASHFREE_APP_ID")
	if clientId == "" {
		log.Fatal("CASHFREE_APP_ID is required")
	}

	clientSecret := os.Getenv("CASHFREE_SECRET_KEY")
	if clientSecret == "" {
		log.Fatal("CASHFREE_SECRET_KEY is required")
	}

	mode := os.Getenv("CASHFREE_MODE")
	if mode == "" {
		mode = "TEST"
	}

	if mode == "PROD" {
		cashfree.XEnvironment = cashfree.PRODUCTION
	} else {
		cashfree.XEnvironment = cashfree.SANDBOX
	}

	// clientId := "TEST398677585a4c6083742542a53d776893"
	// clientSecret := "TEST679577dfe41dc2c61d371bb44ede411fb69c185f"
	cashfree.XClientId = &clientId
	cashfree.XClientSecret = &clientSecret
	cashfree.XEnvironment = cashfree.SANDBOX
	expiryTime := time.Now().Add(10 * time.Minute).Format("2006-01-02T15:04:05Z")
	link, response, err := cashfree.PGCreateLink(
		cashfree.PtrString("2023-08-01"),
		&cashfree.CreateLinkRequest{
			LinkAmount:   100.0,
			LinkCurrency: "INR",
			LinkId:       "TEST24",
			CustomerDetails: cashfree.LinkCustomerDetailsEntity{
				CustomerEmail: cashfree.PtrString("saurav.raj.ash+customer@gmail.com"),
				CustomerPhone: "1234567890",
			},
			LinkMeta: &cashfree.LinkMetaResponseEntity{
				ReturnUrl: cashfree.PtrString("http://192.168.1.4:8080/cart/payment/callback?order_id={order_id}"),
			},
			LinkNotify: &cashfree.LinkNotifyEntity{
				SendEmail: cashfree.PtrBool(true),
				SendSms:   cashfree.PtrBool(true),
			},
			LinkExpiryTime: cashfree.PtrString(expiryTime),
			LinkPurpose:    "Test Order",
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		l.DebugF("Error: %s", err.Error())
		return nil, err
	}
	l.DebugF("Response: %#v", response.StatusCode)
	l.InfoF("Link: %#v", *link.LinkUrl)
	return link, nil
}

type OrderDetail struct {
	OrderId       string  `json:"order_id"`
	OrderAmount   float64 `json:"order_amount"`
	CustomerId    string  `json:"customer_id"`
	CustomerPhone string  `json:"customer_phone"`
	ReturnUrl     string  `json:"return_url"`
}

func createOrder(order OrderDetail) (*cashfree.OrderEntity, error) {
	clientId := "TEST398677585a4c6083742542a53d776893"
	clientSecret := "TEST679577dfe41dc2c61d371bb44ede411fb69c185f"
	cashfree.XClientId = &clientId
	cashfree.XClientSecret = &clientSecret
	cashfree.XEnvironment = cashfree.SANDBOX

	// returnUrl := "http://192.168.1.4:8080/cart/payment/callback?order_id=66910e7311df548fa16d3cc2"

	request := cashfree.CreateOrderRequest{
		OrderAmount:   order.OrderAmount,
		OrderCurrency: "INR",
		OrderId:       &order.OrderId,
		CustomerDetails: cashfree.CustomerDetails{
			CustomerId:    order.CustomerId,
			CustomerPhone: order.CustomerPhone,
		},
		OrderMeta: &cashfree.OrderMeta{
			ReturnUrl: &order.ReturnUrl,
		},
	}

	version := "2023-08-01"

	response, httpResponse, err := cashfree.PGCreateOrder(&version, &request, nil, nil, nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(httpResponse.StatusCode)
		fmt.Println(*response.PaymentSessionId)
	}
	return response, nil
}

// func CreateOrderPayment(app *conf.Config) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var order D
// 	}
// }
