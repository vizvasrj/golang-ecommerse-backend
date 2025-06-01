package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"src/l"
	"src/pkg/conf"
	"time"

	cashfree "github.com/cashfree/cashfree-pg/v4"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
	utils "github.com/razorpay/razorpay-go/utils"
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

func CreateOrder(order OrderDetail) (*cashfree.OrderEntity, error) {
	clientId := os.Getenv("CASHFREE_APP_ID")
	clientSecret := os.Getenv("CASHFREE_SECRET_KEY")
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

// func handleCashFreeWebhook(app *conf.Config) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		clientId := os.Getenv("CASHFREE_APP_ID")
// 		clientSecret := os.Getenv("CASHFREE_SECRET_KEY")
// 		cashfree.XClientId = &clientId
// 		cashfree.XClientSecret = &clientSecret
// 		cashfree.XEnvironment = cashfree.SANDBOX

// 		signature := c.Request.Header.Get("x-webhook-signature")
// 		timestamp := c.Request.Header.Get("x-webhook-timestamp")
// 		l.InfoF("Signature: %s, Timestamp: %s", signature, timestamp)

// 		stringRequestBody, err := io.ReadAll(c.Request.Body)
// 		if err != nil {
// 			l.DebugF("Error reading request body: %s", err.Error())
// 			c.JSON(400, gin.H{"error": "invalid request"})
// 			return
// 		}
// 		l.InfoF("Request body: %s", string(stringRequestBody))
// 		webhookEvent, err := cashfree.PGVerifyWebhookSignature(signature, string(stringRequestBody), timestamp)
// 		if err != nil {
// 			l.DebugF("Error verifying webhook signature: %s", err.Error())
// 			c.JSON(400, gin.H{"error": "invalid request"})
// 			return
// 		}
// 		eventData, err := mapToCashfreeWebhookRequest(webhookEvent)
// 		if err != nil {
// 			l.DebugF("Error mapping webhook event: %s", err.Error())
// 			c.JSON(400, gin.H{"error": "invalid request"})
// 			return
// 		}
// 		if eventData.TxStatus == "SUCCESS" {
// 			// TODO do something with the successful transaction

// 			c.JSON(200, gin.H{"status": "success"})
// 		} else if eventData.TxStatus == "FAILED" {
// 			// TODO do something with the failed transaction
// 			c.JSON(400, gin.H{"status": "failed"})
// 		}
// 	}
// }

func Executerazorpay(amount float64, receptId uuid.UUID, orderId string) (string, map[string]any, error) {
	rzp_id := os.Getenv("RAZORPAY_ID")
	rzp_secret := os.Getenv("RAZORPAY_SECRET")
	l.DebugF("i get order id: %s", orderId)
	client := razorpay.NewClient(rzp_id, rzp_secret)

	// utils.VerifyPaymentSignature(map[string]interface{}{
	// 	"razorpay_order_id":   "order_PRVgVQfRHn6mkP",
	// 	"razorpay_payment_id": "pay_PRVqCzM0SB33Dj",
	// })

	data := map[string]interface{}{
		"amount":   amount * 100,
		"currency": "INR",
		"receipt":  receptId,
		"notes": map[string]interface{}{
			"orderId": orderId,
		},
	}
	l.DebugF("Data: %#v", data)
	body, err := client.Order.Create(data, nil)
	if err != nil {
		return "", nil, errors.New("payment not initiated")
	}

	razorId, _ := body["id"].(string)

	l.DebugF("Razorpay Order ID: %s", razorId)
	return razorId, body, nil
}

// TODO use db to store this
func handleRazorPayWebhook(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rzp_wb_h_secret := "secret"
		eventID := c.Request.Header.Get("x-razorpay-event-id")
		if eventID == "" {
			l.DebugF("Error: Event ID not found")
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		signature := c.Request.Header.Get("X-Razorpay-Signature")
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("Error reading request body:", err.Error())
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		var webhook_data map[string]interface{}
		if utils.VerifyWebhookSignature(string(body), signature, rzp_wb_h_secret) {
			err := json.Unmarshal(body, &webhook_data)
			if err != nil {
				l.DebugF("Error unmarshalling webhook data: %s", err.Error())
			}
		} else {
			log.Println("Error verifying webhook signature")
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		fmt.Printf("Webhook Data: %#v\n", webhook_data)
		if webhook_data["event"].(string) == "order.paid" {
			receiptId := webhook_data["payload"].(map[string]interface{})["order"].(map[string]interface{})["entity"].(map[string]interface{})["receipt"].(string)
			receiptIdObject, err := uuid.Parse(receiptId)
			if err != nil {
				l.DebugF("Error converting receipt id to UUID: %s", err.Error())
				c.JSON(400, gin.H{"error": "invalid request"})
				return
			}

			// Fetch receipt from PostgreSQL
			receiptDoc := Receipt{}
			query := `SELECT id, order_id, provider_data FROM receipts WHERE id = $1`
			row := app.DB.QueryRowContext(c, query, receiptIdObject)

			var providerData []byte
			err = row.Scan(&receiptDoc.ID, &receiptDoc.OrderID, &providerData)
			if err != nil {
				l.DebugF("Error fetching receipt: %s", err.Error())
			}

			if err := json.Unmarshal(providerData, &receiptDoc.ProviderData); err != nil {
				l.DebugF("Error unmarshaling provider data: %s", err.Error())
			}

			razorpayOrderID, _ := receiptDoc.ProviderData["razorpay_order_id"].(string)
			paymentId := webhook_data["payload"].(map[string]interface{})["payment"].(map[string]interface{})["entity"].(map[string]interface{})["id"].(string)

			// Verify payment signature
			ok := utils.VerifyPaymentSignature(map[string]interface{}{
				"razorpay_order_id":   razorpayOrderID,
				"razorpay_payment_id": paymentId,
			}, signature, rzp_wb_h_secret)
			if !ok {
				l.DebugF("Error verifying payment signature")
				c.JSON(400, gin.H{"error": "invalid request"})
				return
			}
		}

		// Update order status
		notes := webhook_data["payload"].(map[string]interface{})["payment"].(map[string]interface{})["entity"].(map[string]interface{})["notes"]
		orderId := ""
		if noteValues, ok := notes.(map[string]interface{}); ok {
			orderId = noteValues["orderId"].(string)
		}

		orderUUID, err := uuid.Parse(orderId)
		if err != nil {
			l.DebugF("Error converting order id to UUID: %s", err.Error())
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		// Convert webhook data to JSONB
		providerDataJSON, err := json.Marshal(webhook_data)
		if err != nil {
			l.DebugF("Error marshaling provider data: %s", err.Error())
			c.JSON(500, gin.H{"error": "internal server error"})
			return
		}

		// Update receipt in PostgreSQL
		updateQuery := `
            UPDATE receipts
            SET payment_status = $1,
                updated = $2,
                provider_data = $3
            WHERE order_id = $4
        `
		_, err = app.DB.ExecContext(c, updateQuery,
			PaymentStatusCaptured,
			time.Now().UTC(),
			providerDataJSON,
			orderUUID,
		)
		if err != nil {
			l.DebugF("Error updating receipt: %s", err.Error())
			c.JSON(500, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	}
}

// func verifyPaymentSignature(data map[string]string, signature string) err {
// 	rzp_id := os.Getenv("RAZORPAY_ID")
// 	rzp_secret := os.Getenv("RAZORPAY_SECRET")
// 	webhook_secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET_KEY")
// 	client := razorpay.NewClient(rzp_id, rzp_secret)
// 	err := utils.VerifyPaymentSignature(data, signature, webhook_secret)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func unmarshalWebhookEvent(data map[string]interface{}) (*razorpay.WebhookEvent, error) {
// 	event := &razorpay.WebhookEvent{}
// 	err := utils.Unmarshal(data, event)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return event, nil
// }

func getOrderIdAndPaymentId(webhook_data map[string]interface{}) (string, string) {
	// Extract order_id and payment_id from webhook_data
	payload, ok := webhook_data["payload"].(map[string]interface{})
	if !ok {
		return "", ""
	}

	payment, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return "", ""
	}

	entity, ok := payment["entity"].(map[string]interface{})
	if !ok {
		return "", ""
	}

	orderId, _ := entity["order_id"].(string)
	paymentId, _ := entity["id"].(string)

	return orderId, paymentId
}

func saveJsonWebhookData(w interface{}) {
	// Save the webhook data to a file
	file, _ := json.MarshalIndent(w, "", " ")
	_ = os.WriteFile("/tmp/webhook_data.json", file, 0644)
}
