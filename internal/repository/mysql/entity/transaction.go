package entity

import (
	"database/sql" // Untuk sql.NullString jika description bisa NULL
	"time"
)

// TransactionType merepresentasikan tipe transaksi (income atau expense).
type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

// Transaction merepresentasikan entitas transaksi di database.
type Transaction struct {
	ID              int64           `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          int64           `gorm:"column:user_id"`
	CategoryID      sql.NullInt64   `gorm:"column:category_id"` 
	Amount          float64         `gorm:"column:amount;type:decimal(15,2)"` 
	Type            TransactionType `gorm:"column:type"`                     
	Description     sql.NullString  `gorm:"column:description"`           
	TransactionDate time.Time       `gorm:"column:transaction_date"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
}

// TableName mengembalikan nama tabel di database untuk model Transaction.
func (Transaction) TableName() string {
	return "transactions"
}