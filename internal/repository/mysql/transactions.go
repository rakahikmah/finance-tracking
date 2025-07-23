package mysql

import (
	"context"
	"database/sql" 
	"github.com/rakahikmah/finance-tracking/config"
	"github.com/rakahikmah/finance-tracking/internal/helper"
 	"github.com/rakahikmah/finance-tracking/internal/repository/mysql/entity" 
	apperr "github.com/rakahikmah/finance-tracking/error"

	errwrap "github.com/pkg/errors"
	"gorm.io/gorm"
)


type TransactionWithCategory struct {
	entity.Transaction            
	CategoryName sql.NullString `gorm:"column:category_name"` 
}

// TransactionSummaryByCategory adalah struct untuk menampung hasil ringkasan per kategori dan tipe.
type TransactionSummaryByCategory struct {
	CategoryName sql.NullString `gorm:"column:category_name"`
	Type         string         `gorm:"column:type"`
	TotalAmount  float64        `gorm:"column:total_amount"`
}

// ITransactionRepository mendefinisikan interface untuk operasi CRUD pada entitas Transaction.
type ITransactionRepository interface {
	TrxSupportRepo // Warisan dari interface transaksi (biasanya ada di file mysql/common.go)

	
	GetByIDAndUserID(ctx context.Context, ID int64, userID int64) (e *entity.Transaction, err error)

	Create(ctx context.Context, dbTrx TrxObj, params *entity.Transaction, nonZeroVal bool) error
	Update(ctx context.Context, dbTrx TrxObj, params *entity.Transaction, changes *entity.Transaction) (err error)
	DeleteByIDAndUserID(ctx context.Context, dbTrx TrxObj, id int64, userID int64) error
	GetAllByUserID(ctx context.Context, userID int64) (result []*TransactionWithCategory, err error)
	GetSummaryByCategoryAndTypeByUserID(ctx context.Context, userID int64, startDate, endDate string) (result []*TransactionSummaryByCategory, err error)
	GetDailySummaryByUserID(ctx context.Context, userID int64, startDate, endDate string) (result []map[string]interface{}, err error)
}

// TransactionRepository adalah implementasi repository untuk entitas Transaction.
type TransactionRepository struct {
	GormTrxSupport // Warisan dari struct untuk dukungan transaksi
}

// NewTransactionRepository membuat instance baru dari TransactionRepository.
func NewTransactionRepository(mysql *config.Mysql) *TransactionRepository {
	return &TransactionRepository{GormTrxSupport{db: mysql.DB}}
}



// GetAllByUserID mengambil semua transaksi yang dimiliki oleh user tertentu, termasuk nama kategori.
func (r *TransactionRepository) GetAllByUserID(ctx context.Context, userID int64) (result []*TransactionWithCategory, err error) {
	funcName := "TransactionRepository.GetAllByUserID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	// Menggunakan Raw SQL untuk JOIN dan mengambil category_name
	// Pastikan alias kolom `c.name` menjadi `category_name` agar cocok dengan TransactionWithCategory.
	// Jika category_id adalah NULL, c.name juga akan NULL (LEFT JOIN).
	query := `
		SELECT
			t.id, t.user_id, t.category_id, t.amount, t.type, t.description, t.transaction_date, t.created_at, t.updated_at,
			c.name as category_name
		FROM
			transactions t
		LEFT JOIN
			categories c ON t.category_id = c.id
		WHERE
			t.user_id = ?
		ORDER BY
			t.transaction_date DESC, t.id DESC
	`
	err = r.db.Raw(query, userID).Scan(&result).Error
	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		return []*TransactionWithCategory{}, nil // Mengembalikan slice kosong jika tidak ada record
	}
	if err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	return result, nil
}

// GetByIDAndUserID mengambil transaksi berdasarkan ID dan user ID-nya.
// Ini penting untuk otorisasi agar user hanya bisa melihat/memodifikasi transaksinya sendiri.
// Mengembalikan *entity.Transaction karena tidak selalu perlu nama kategori di sini.
func (r *TransactionRepository) GetByIDAndUserID(ctx context.Context, ID int64, userID int64) (result *entity.Transaction, err error) {
	funcName := "TransactionRepository.GetByIDAndUserID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	// Wajib menambahkan filter WHERE user_id = ? untuk keamanan!
	err = r.db.Where("id = ? AND user_id = ?", ID, userID).First(&result).Error
	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrRecordNotFound()
	}
	if err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	return result, nil
}

// GetDailySummaryByUserID contoh fungsi untuk mendapatkan ringkasan transaksi per hari untuk user tertentu.
// Ini bisa dikembangkan lebih lanjut (misal: filter type, category, etc.)
func (r *TransactionRepository) GetDailySummaryByUserID(ctx context.Context, userID int64, startDate, endDate string) (result []map[string]interface{}, err error) {
	funcName := "TransactionRepository.GetDailySummaryByUserID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	// Contoh SQL untuk ringkasan harian
	// Sum amount by transaction_date and type, grouped by user_id
	err = r.db.Raw(`
		SELECT
			DATE(transaction_date) as transaction_day,
			type,
			SUM(amount) as total_amount
		FROM
			transactions
		WHERE
			user_id = ? AND transaction_date BETWEEN ? AND ?
		GROUP BY
			transaction_day, type
		ORDER BY
			transaction_day ASC, type ASC
	`, userID, startDate, endDate).Scan(&result).Error

	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		return []map[string]interface{}{}, nil // Mengembalikan slice kosong jika tidak ada record
	}
	if err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}
	return result, nil
}

// Create membuat transaksi baru.
func (r *TransactionRepository) Create(ctx context.Context, dbTrx TrxObj, params *entity.Transaction, nonZeroVal bool) error {
	funcName := "TransactionRepository.Create"

	if err := helper.CheckDeadline(ctx); err != nil {
		return errwrap.Wrap(err, funcName)
	}

	cols := helper.NonZeroCols(params, nonZeroVal)
	return r.Trx(dbTrx).Select(cols).Create(&params).Error
}

// Update memperbarui transaksi yang ada.
// Wajib menambahkan filter user_id untuk otorisasi.
func (r *TransactionRepository) Update(ctx context.Context, dbTrx TrxObj, params *entity.Transaction, changes *entity.Transaction) error {
	funcName := "TransactionRepository.Update"

	if err := helper.CheckDeadline(ctx); err != nil {
		return errwrap.Wrap(err, funcName)
	}

	if params.ID == 0 || params.UserID == 0 {
		return errwrap.Wrap(apperr.ErrInvalidRequest().SetDetail("Transaction ID or User ID is missing."), funcName)
	}

	db := r.Trx(dbTrx).Model(params).Where("user_id = ?", params.UserID)

	var err error
	if changes != nil {
		err = db.Updates(*changes).Error
	} else {
		err = db.Updates(helper.StructToMap(params, false)).Error
	}

	if err != nil {
		return errwrap.Wrap(err, funcName)
	}

	return nil
}

// DeleteByIDAndUserID menghapus transaksi berdasarkan ID dan user ID-nya.
// Wajib menambahkan filter user_id untuk otorisasi.
func (r *TransactionRepository) DeleteByIDAndUserID(ctx context.Context, dbTrx TrxObj, id int64, userID int64) error {
	funcName := "TransactionRepository.DeleteByIDAndUserID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return errwrap.Wrap(err, funcName)
	}

	if userID == 0 {
		return errwrap.Wrap(apperr.ErrInvalidRequest().SetDetail("User ID is missing for delete operation."), funcName)
	}

	err := r.Trx(dbTrx).Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Transaction{}).Error
	if err != nil {
		return errwrap.Wrap(err, funcName)
	}

	return nil
}


func (r *TransactionRepository) GetSummaryByCategoryAndTypeByUserID(ctx context.Context, userID int64, startDate, endDate string) (result []*TransactionSummaryByCategory, err error) {
	funcName := "TransactionRepository.GetSummaryByCategoryAndTypeByUserID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	query := `
		SELECT
			COALESCE(c.name, 'Uncategorized') as category_name, -- Gunakan COALESCE untuk kategori NULL
			t.type,
			SUM(t.amount) as total_amount
		FROM
			transactions t
		LEFT JOIN
			categories c ON t.category_id = c.id
		WHERE
			t.user_id = ? AND t.transaction_date BETWEEN ? AND ?
		GROUP BY
			category_name, t.type
		ORDER BY
			category_name ASC, t.type ASC
	`
	err = r.db.Raw(query, userID, startDate, endDate).Scan(&result).Error

	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		return []*TransactionSummaryByCategory{}, nil // Mengembalikan slice kosong jika tidak ada record
	}
	if err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}
	return result, nil
}