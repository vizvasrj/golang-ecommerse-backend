package payment

import (
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashfreeWebhookRequest struct {
	OrderId     string `json:"orderId"`
	OrderAmount string `json:"orderAmount"`
	ReferenceId string `json:"referenceId"`
	TxStatus    string `json:"txStatus"`
	PaymentMode string `json:"paymentMode"`
	TxMsg       string `json:"txMsg"`
	TxTime      string `json:"txTime"`
	Signature   string `json:"signature"`
}

func mapToCashfreeWebhookRequest(data interface{}) (*CashfreeWebhookRequest, error) {
	request := CashfreeWebhookRequest{}
	err := json.Unmarshal(data.([]byte), &request)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling data: %v", err)
	}

	return &request, nil
}

type RazorpayWebhookEntity struct {
	Entity    string   `json:"entity"`
	AccountID string   `json:"account_id"`
	Event     string   `json:"event"`
	Contains  []string `json:"contains"`
	Payload   struct {
		Payment struct {
			Entity struct {
				ID               string        `json:"id"`
				Entity           string        `json:"entity"`
				Amount           int           `json:"amount"`
				Currency         string        `json:"currency"`
				Status           string        `json:"status"`
				OrderID          string        `json:"order_id"`
				InvoiceID        interface{}   `json:"invoice_id"`
				International    bool          `json:"international"`
				Method           string        `json:"method"`
				AmountRefunded   int           `json:"amount_refunded"`
				RefundStatus     interface{}   `json:"refund_status"`
				Captured         bool          `json:"captured"`
				Description      string        `json:"description"`
				CardID           interface{}   `json:"card_id"`
				Bank             interface{}   `json:"bank"`
				Wallet           interface{}   `json:"wallet"`
				Vpa              string        `json:"vpa"`
				Email            string        `json:"email"`
				Contact          string        `json:"contact"`
				Notes            []interface{} `json:"notes"`
				Fee              int           `json:"fee"`
				Tax              int           `json:"tax"`
				ErrorCode        interface{}   `json:"error_code"`
				ErrorDescription interface{}   `json:"error_description"`
				ErrorSource      interface{}   `json:"error_source"`
				ErrorStep        interface{}   `json:"error_step"`
				ErrorReason      interface{}   `json:"error_reason"`
				AcquirerData     struct {
					Rrn              string `json:"rrn"`
					UpiTransactionID string `json:"upi_transaction_id"`
				} `json:"acquirer_data"`
				CreatedAt int         `json:"created"`
				Reward    interface{} `json:"reward"`
				Upi       struct {
					Vpa string `json:"vpa"`
				} `json:"upi"`
				BaseAmount int `json:"base_amount"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
	CreatedAt int `json:"created"`
}

type PaymentStatus string

const (
	PaymentStatusCaptured PaymentStatus = "captured"
	PaymentStatusFailed   PaymentStatus = "failed"
)

type Receipt struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id"`
	OrderID         primitive.ObjectID `json:"orderId" bson:"orderId"`
	CartID          primitive.ObjectID `json:"cartId" bson:"cartId"`
	RazorpayOrderID string             `json:"razorpay_order_id" bson:"razorpay_order_id"`
	Amount          float64            `json:"amount" bson:"amount"`
	Created         time.Time          `json:"created" bson:"created"`
	Updated         time.Time          `json:"updated" bson:"updated"`
	PaymentProvider string             `json:"paymentProvider" bson:"paymentProvider"`
	ProviderData    interface{}        `json:"providerData" bson:"providerData"`
	PaymentStatus   PaymentStatus      `json:"paymentStatus" bson:"paymentStatus"`
}

type OrderCreateRequest struct {
	Amount                int               `json:"amount"`
	Currency              string            `json:"currency"`
	Receipt               string            `json:"receipt"`
	PartialPayment        bool              `json:"partial_payment"`
	Notes                 map[string]string `json:"notes,omitempty"`
	FirstPaymentMinAmount int               `json:"first_payment_min_amount,omitempty"`
}
type OrderCreateResponse struct {
	ID         string      `json:"id"`
	Entity     string      `json:"entity"`
	Amount     uint        `json:"amount"`
	AmountPaid uint        `json:"amount_paid,omitempty"`
	AmountDue  uint        `json:"amount_due,omitempty"`
	Currency   string      `json:"currency"`
	Receipt    string      `json:"receipt"`
	OfferID    interface{} `json:"offer_id,omitempty"`
	Status     string      `json:"status"`
	Attempts   int         `json:"attempts"`
	Notes      interface{} `json:"notes"`
	CreatedAt  int         `json:"created"`
}

func (a *OrderCreateResponse) GetNotes() map[string]string {
	switch noteValues := a.Notes.(type) {
	case map[string]interface{}:
		notes := make(map[string]string)
		for key, value := range noteValues {
			notes[key] = value.(string)
		}
		return notes
	case []interface{}:
		return nil
	}
	return nil
}

func (a *OrderCreateResponse) CreatedTime() time.Time {
	return time.Unix(int64(a.CreatedAt), 0)
}
