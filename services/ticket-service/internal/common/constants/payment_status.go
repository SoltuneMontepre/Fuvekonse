package constants

// PaymentStatus represents the status of a payment
type PaymentStatus string

// Payment status constants
const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
)

// String returns the string representation of the payment status
func (ps PaymentStatus) String() string {
	return string(ps)
}

// IsValid checks if the payment status is valid
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusPaid, PaymentStatusCancelled,
		PaymentStatusFailed, PaymentStatusExpired:
		return true
	default:
		return false
	}
}

// CanCancel checks if a payment with this status can be cancelled
func (ps PaymentStatus) CanCancel() bool {
	return ps == PaymentStatusPending
}


// IsCompleted checks if the payment process is completed
func (ps PaymentStatus) IsCompleted() bool {
	return ps == PaymentStatusPaid || ps == PaymentStatusCancelled ||
		ps == PaymentStatusFailed || ps == PaymentStatusExpired
}
