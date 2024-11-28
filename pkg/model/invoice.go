package model

import (
	"fmt"
	"github.com/Rhymond/go-money"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/utils"
	"github.com/e-harsley/numerisbook_test/pkg"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type (
	InvoiceConfiguration struct {
		//This holds configurations such as 7days reminder for invoice
		ID               primitive.ObjectID            `json:"_id" bson:"_id"`
		CreatedAt        *time.Time                    `json:"created_at" bson:"created_at"`
		UpdatedAt        *time.Time                    `json:"updated_at" bson:"updated_at"`
		Name             string                        `json:"name" bson:"name"` // 14days reminder
		Active           bool                          `json:"active" bson:"active"`
		ReminderDuration pkg.ReminderConfigurationType `json:"reminder_duration"`
		ReminderValue    int64                         `json:"reminder_value"`
		UserID           primitive.ObjectID            `json:"user_id" bson:"user_id" lookup:"user:user_id:_id"`
		User             User                          `json:"user" bson:"-"`
	}

	Invoice struct {
		ID                 primitive.ObjectID `json:"_id" bson:"_id"`
		CreatedAt          *time.Time         `json:"created_at" bson:"created_at"`
		UpdatedAt          *time.Time         `json:"updated_at" bson:"updated_at"`
		Name               string             `json:"name" bson:"name"`
		InvoiceID          string             `json:"invoice_id" bson:"invoice_id"`
		DateIssued         time.Time          `json:"date_issued" bson:"date_issued"`
		DueDate            time.Time          `json:"due_date" bson:"due_date"`
		Currency           string             `json:"currency" bson:"currency"`
		Items              []Item             `json:"items" bson:"items"`
		UserID             primitive.ObjectID `json:"user_id" bson:"user_id" lookup:"user:user_id:_id:user:struct"`
		User               User               `json:"user" bson:"-"`
		DiscountType       pkg.DiscountType   `json:"discount_type" bson:"discount_type"`
		DiscountValue      float64            `json:"discount_value" bson:"discount_value"`
		SubTotal           float64            `json:"sub_total" bson:"sub_total"`
		AmountDue          float64            `json:"amount_due" bson:"amount_due"`
		DiscountAmount     float64            `json:"discount_amount" bson:"discount_amount"`
		PaymentInformation PaymentInformation `json:"payment_information"  bson:"payment_information"`
		Note               string             `json:"note" bson:"note"`
		PaymentStatus      pkg.PaymentStatus  `json:"payment_status" bson:"payment_status"`
		Status             pkg.InvoiceStatus  `json:"status" bson:"status"`
	}

	InvoiceReminder struct {
		ID              primitive.ObjectID `json:"_id" bson:"_id"`
		CreatedAt       *time.Time         `json:"created_at" bson:"created_at"`
		UpdatedAt       *time.Time         `json:"updated_at" bson:"updated_at"`
		UserID          primitive.ObjectID `json:"user_id" bson:"user_id" lookup:"user:user_id:_id"`
		User            User               `json:"user" bson:"-"`
		InvoiceID       primitive.ObjectID `json:"invoice_id" bson:"invoice_id" lookup:"invoice:invoice_id:_id"`
		Invoice         Invoice            `json:"invoice" bson:"-"`
		TimeForReminder time.Time          `json:"time_for_reminder" bson:"time_for_reminder"`
		Active          bool               `json:"active" bson:"active"`
	}

	Item struct {
		Name        string  `json:"name" bson:"name"`
		Description string  `json:"description" bson:"description"`
		Quantity    int64   `json:"quantity" bson:"quantity"`
		Value       float64 `json:"value" bson:"value"`
		Currency    string  `json:"currency" bson:"currency"`
		Total       float64 `json:"total" bson:"total"`
	}

	PaymentInformation struct {
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
		AchRoutingNo  string `json:"ach_routing_no"`
		BankName      string `json:"bank_name"`
		BankAddress   string `json:"bank_address"`
	}

	InvoiceActivityLog struct {
		ID        primitive.ObjectID    `json:"_id" bson:"_id"`
		CreatedAt *time.Time            `json:"created_at" bson:"created_at"`
		UpdatedAt *time.Time            `json:"updated_at" bson:"updated_at"`
		UserID    primitive.ObjectID    `json:"user_id" bson:"user_id" lookup:"user:user_id:_id:user:struct"`
		User      User                  `json:"user" bson:"-"`
		Action    pkg.ActivityLogAction `json:"action" bson:"action"`
		Message   string                `json:"message" bson:"message"`
		InvoiceID primitive.ObjectID    `json:"invoice_id" bson:"invoice_id" lookup:"invoice:invoice_id:_id:invoice:struct"`
		Invoice   Invoice               `json:"invoice" bson:"-"`
	}
)

func (cls InvoiceConfiguration) GetModelName() string {
	return "invoice_configuration"
}

func (cls Invoice) GetModelName() string {
	return "invoice"
}

func (cls InvoiceActivityLog) GetModelName() string {
	return "invoice_activity_log"
}

func (cls InvoiceReminder) GetModelName() string {
	return "invoice_reminder"
}

func (cls *Invoice) SetDefaultInvoice() {

	cls.InvoiceID = utils.GenerateNumericID(9, 5, "INV")
	cls.DateIssued = time.Now()
	cls.Status = pkg.InvoicePending
	cls.PaymentStatus = pkg.Pending

}

func (ic *InvoiceConfiguration) GetReminderTriggerDate() time.Time {
	now := time.Now()

	switch ic.ReminderDuration {
	case pkg.Days:
		return now.AddDate(0, 0, int(ic.ReminderValue))
	case pkg.Hours:
		return now.Add(-time.Duration(ic.ReminderValue) * time.Hour)
	case pkg.Months:
		return now.AddDate(0, int(ic.ReminderValue), 0)
	default:
		return now.AddDate(0, 0, int(ic.ReminderValue))
	}
}

func (cls *Invoice) CalculatePayment() {
	var totalMoney *money.Money
	totalMoney = money.New(0, cls.Currency)
	var discountMoney *money.Money
	discountMoney = money.New(0, cls.Currency)

	for i := range cls.Items {
		itemTotal := float64(cls.Items[i].Quantity) * cls.Items[i].Value
		itemMoney := money.New(int64(itemTotal*100), cls.Items[i].Currency)
		totalMoney, _ = totalMoney.Add(itemMoney)
		cls.Items[i].Total = itemTotal
	}
	cls.SubTotal = float64(totalMoney.Amount()) / 100.0
	if cls.DiscountValue > 0 {
		switch cls.DiscountType {
		case pkg.Percentage:
			discountAmount := float64(totalMoney.Amount()/100) * (cls.DiscountValue / 100)
			discountMoney = money.New(int64(discountAmount*100), cls.Currency)
			totalMoney, _ = totalMoney.Subtract(discountMoney)
		case pkg.FixedAmount:
			discountMoney = money.New(int64(cls.DiscountValue*100), cls.Currency)
			totalMoney, _ = totalMoney.Subtract(discountMoney)
		}
	}

	cls.AmountDue = float64(totalMoney.Amount()) / 100.0
	cls.DiscountAmount = float64(discountMoney.Amount()) / 100.0
	fmt.Println(cls.AmountDue, cls.AmountDue)
}
