// internal/usecase/transactions/entity/crud.go

package entity



// TransactionTypeString dan konstanta tetap sama
type TransactionTypeString string

const (
	TransactionTypeIncomeStr  TransactionTypeString = "income"
	TransactionTypeExpenseStr TransactionTypeString = "expense"
)

// TransactionReq tetap sama
type TransactionReq struct {
	UserID          int64                 `json:"user_id,omitempty"`
	CategoryID      *int64                `json:"category_id"`
	Amount          float64               `json:"amount" validate:"required,gt=0" name:"Jumlah Transaksi"`
	Type            TransactionTypeString `json:"type" validate:"required,oneof=income expense" name:"Tipe Transaksi"`
	Description     *string               `json:"description"`
	TransactionDate string                `json:"transaction_date" validate:"required,datetime=2006-01-02" name:"Tanggal Transaksi"`
}

// TransactionResponse adalah struktur data untuk output (response body) saat mengembalikan data transaksi.
type TransactionResponse struct {
	ID              int64                 `json:"id"`
	UserID          int64                 `json:"user_id"`
	CategoryID      *int64                `json:"category_id"`
	CategoryName    *string               `json:"category_name"` 
	Amount          float64               `json:"amount"`
	Type            TransactionTypeString `json:"type"`
	Description     *string               `json:"description"`
	TransactionDate string                `json:"transaction_date"`
	CreatedAt       string                `json:"created_at"`
	UpdatedAt       string                `json:"updated_at"`
}

// TransactionSummaryResponse adalah struktur data untuk respons ringkasan transaksi per kategori dan tipe.
type TransactionSummaryResponse struct {
	CategoryName *string               `json:"category_name"`
	Type         TransactionTypeString `json:"type"`
	TotalAmount  float64               `json:"total_amount"`
}

// SetUserID method tetap sama
func (r *TransactionReq) SetUserID(userID int64) {
	r.UserID = userID
}