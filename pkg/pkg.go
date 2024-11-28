package pkg

type DiscountType string

const (
	Percentage  DiscountType = "percentage"
	FixedAmount DiscountType = "fixed_amount"
)

type ReminderConfigurationType string

const (
	Days   ReminderConfigurationType = "days"
	Hours  ReminderConfigurationType = "hours"
	Months ReminderConfigurationType = "months"
)

type PaymentStatus string

const (
	PartPaid PaymentStatus = "part_paid"
	Paid     PaymentStatus = "paid"
	Pending  PaymentStatus = "pending"
	NotPaid  PaymentStatus = "not_paid"
)

type InvoiceStatus string

const (
	Drafted        InvoiceStatus = "drafted"
	OverDue        InvoiceStatus = "overdue"
	InvoicePaid    InvoiceStatus = "paid"
	InvoiceNotPaid InvoiceStatus = "not_paid"
	InvoicePending InvoiceStatus = "pending"
)

type ActivityLogAction string

const (
	InvoiceCreated          ActivityLogAction = "invoice_created"
	InvoiceUpdated          ActivityLogAction = "invoice_updated"
	ActivityLogPaid         ActivityLogAction = "invoice_paid"
	InvoiceNotificationSent ActivityLogAction = "invoice_notification_sent"
)
