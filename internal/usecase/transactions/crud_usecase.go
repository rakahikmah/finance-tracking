package transactions_usecase // Nama paket

import (
	"context"
	"database/sql" // Untuk sql.NullInt64 dan sql.NullString
	"errors"
	"fmt"
	"strconv"
	"time" // Untuk time.Time, time.Parse, dan DatetimeNowJakarta

	generalEntity "github.com/rakahikmah/finance-tracking/entity" // Asumsi ini entity dasar seperti CaptureFields
	"github.com/rakahikmah/finance-tracking/internal/helper"
	"github.com/rakahikmah/finance-tracking/internal/repository/mysql"
	myentity "github.com/rakahikmah/finance-tracking/internal/repository/mysql/entity" // Model GORM Transaction
	usecaseEntity "github.com/rakahikmah/finance-tracking/internal/usecase/transactions/entity" // DTO TransactionReq/Response

	apperr "github.com/rakahikmah/finance-tracking/error" // Jika ada error kustom dari project Anda
)

// CrudTransaction adalah struct yang akan menampung dependensi repository.
type CrudTransaction struct {
	TransactionRepo mysql.ITransactionRepository // Menggunakan interface repository Transaction
	CategoryRepo    mysql.ICategoryRepository    // Perlu untuk validasi category_id
}

// NewCrudTransaction adalah konstruktor untuk CrudTransaction.
func NewCrudTransaction(
	TransactionRepo mysql.ITransactionRepository,
	CategoryRepo mysql.ICategoryRepository, // Tambahkan CategoryRepo
) *CrudTransaction {
	return &CrudTransaction{
		TransactionRepo: TransactionRepo,
		CategoryRepo:    CategoryRepo,
	}
}

// ICrudTransaction mendefinisikan interface untuk operasi CRUD pada Transaction.
type ICrudTransaction interface {
	Create(ctx context.Context, userID int64, req usecaseEntity.TransactionReq) error
	GetAll(ctx context.Context, userID int64) ([]usecaseEntity.TransactionResponse, error)
	Update(ctx context.Context, id int64, userID int64, req usecaseEntity.TransactionReq) error
	Delete(ctx context.Context, id int64, userID int64) error
	GetDailySummary(ctx context.Context, userID int64, startDate, endDate string) ([]map[string]interface{}, error) // Contoh API tambahan
	GetSummaryByCategoryAndType(ctx context.Context, userID int64, startDate, endDate string) ([]usecaseEntity.TransactionSummaryResponse, error)
}


// Create membuat transaksi baru untuk user tertentu.
func (u *CrudTransaction) Create(ctx context.Context, userID int64, req usecaseEntity.TransactionReq) error {
	funcName := "CrudTransaction.Create"

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, nil, "UserID tidak ditemukan")
		return apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"type":    string(req.Type),
		"amount":  fmt.Sprintf("%.2f", req.Amount),
	}

	// Validasi CategoryID jika diberikan
	var categoryID sql.NullInt64
	if req.CategoryID != nil {
		if *req.CategoryID > 0 {
			// Periksa apakah category_id yang diberikan valid dan milik user yang sama
			category, err := u.CategoryRepo.GetByID(ctx, *req.CategoryID)
			if err != nil {
				helper.LogError(funcName, "CategoryRepo.GetByID", err, logFields, "Error getting category for transaction")
				return apperr.ErrInvalidRequest().SetDetail("Invalid Category ID provided.")
			}
			// Pastikan kategori yang dipilih milik user yang sedang login
			if category.CreatedBy != userID {
				helper.LogError(funcName, "CategoryRepo.GetByID", errors.New("unauthorized category access"), logFields, "User tried to use category not owned by them")
				return apperr.ErrUnauthorized().SetDetail("You are not authorized to use this category.")
			}
			categoryID.Int64 = *req.CategoryID
			categoryID.Valid = true
		}
	}

	// Parse TransactionDate
	parsedDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		helper.LogError(funcName, "time.Parse", err, logFields, "Invalid Transaction Date format")
		return apperr.ErrInvalidRequest().SetDetail("Invalid transaction_date format. Use YYYY-MM-DD.")
	}

	data := &myentity.Transaction{
		UserID:          userID, // Diisi dari parameter yang aman
		CategoryID:      categoryID,
		Amount:          req.Amount,
		Type:            myentity.TransactionType(req.Type), // Konversi ke tipe ENUM Go
		Description:     sql.NullString{String: *req.Description, Valid: req.Description != nil}, // Handle nil pointer for description
		TransactionDate: parsedDate,
		CreatedAt:       helper.DatetimeNowJakarta(), // Menggunakan helper
		UpdatedAt:       helper.DatetimeNowJakarta(), // Menggunakan helper
	}

	// Panggil repository untuk membuat record
	err = u.TransactionRepo.Create(ctx, nil, data, false)
	if err != nil {
		helper.LogError(funcName, "TransactionRepo.Create", err, logFields, "")
		return err
	}

	return nil
}

// GetAll mengambil semua transaksi untuk user tertentu.
func (u *CrudTransaction) GetAll(ctx context.Context, userID int64) ([]usecaseEntity.TransactionResponse, error) {
	funcName := "CrudTransaction.GetAll"
	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"layer":   "usecase",
	}

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return nil, apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// Ambil data dari repository, yang sekarang mengembalikan TransactionWithCategory
	data, err := u.TransactionRepo.GetAllByUserID(ctx, userID) // Ini akan mengembalikan []*mysql.TransactionWithCategory
	if err != nil {
		helper.LogError(funcName, "TransactionRepo.GetAllByUserID", err, logFields, "")
		return nil, err
	}

	// Mapping ke response DTO
	var result []usecaseEntity.TransactionResponse
	for _, row := range data { // `row` sekarang adalah *mysql.TransactionWithCategory
		// Konversi sql.NullInt64/NullString ke pointer atau nilai default
		var categoryID *int64
		if row.CategoryID.Valid {
			categoryID = &row.CategoryID.Int64
		}
		var description *string
		if row.Description.Valid {
			description = &row.Description.String
		}
		var categoryName *string // Handle CategoryName dari TransactionWithCategory
		if row.CategoryName.Valid {
			categoryName = &row.CategoryName.String
		}

		result = append(result, usecaseEntity.TransactionResponse{
			ID:              row.ID,
			UserID:          row.UserID,
			CategoryID:      categoryID,
			CategoryName:    categoryName, // MAP FIELD BARU INI
			Amount:          row.Amount,
			Type:            usecaseEntity.TransactionTypeString(row.Type),
			Description:     description,
			TransactionDate: row.TransactionDate.Format("2006-01-02"),       // Format ke YYYY-MM-DD
			CreatedAt:       helper.ConvertToJakartaTime(row.CreatedAt), // Menggunakan helper
			UpdatedAt:       helper.ConvertToJakartaTime(row.UpdatedAt), // Menggunakan helper
		})
	}

	return result, nil
}

// Update memperbarui transaksi berdasarkan ID dan memastikan milik user yang benar.
func (u *CrudTransaction) Update(ctx context.Context, id int64, userID int64, req usecaseEntity.TransactionReq) error {
	funcName := "CrudTransaction.Update"
	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"id":      fmt.Sprintf("%d", id),
	}

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// 1. Ambil data lama dari database (melibatkan otorisasi user_id)
	oldData, err := u.TransactionRepo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		helper.LogError(funcName, "GetByIDAndUserID", err, logFields, "Error getting existing transaction for update")
		return err // Error akan berupa ErrRecordNotFound atau error lain dari repo
	}

	// 2. Validasi CategoryID jika diubah
	var newCategoryID sql.NullInt64
	if req.CategoryID != nil {
		if *req.CategoryID > 0 {
			category, err := u.CategoryRepo.GetByID(ctx, *req.CategoryID)
			if err != nil {
				helper.LogError(funcName, "CategoryRepo.GetByID", err, logFields, "Invalid Category ID provided for update.")
				return apperr.ErrInvalidRequest().SetDetail("Invalid Category ID provided for update.")
			}
			if category.CreatedBy != userID {
				helper.LogError(funcName, "CategoryRepo.GetByID", errors.New("unauthorized category access"), logFields, "User tried to use category not owned by them for update")
				return apperr.ErrUnauthorized().SetDetail("You are not authorized to use this category for update.")
			}
			newCategoryID.Int64 = *req.CategoryID
			newCategoryID.Valid = true
		}
	} else { // Jika CategoryID di request adalah nil, set menjadi NULL di DB
		newCategoryID.Valid = false
	}


	// Parse TransactionDate jika diubah
	var parsedDate time.Time
	if req.TransactionDate != "" {
		parsedDate, err = time.Parse("2006-01-02", req.TransactionDate)
		if err != nil {
			helper.LogError(funcName, "time.Parse", err, logFields, "Invalid Transaction Date format for update")
			return apperr.ErrInvalidRequest().SetDetail("Invalid transaction_date format. Use YYYY-MM-DD.")
		}
	} else {
        // Jika transaction_date tidak diubah, pertahankan yang lama dari oldData
        parsedDate = oldData.TransactionDate
    }

	// Siapkan perubahan data (hanya field yang diubah)
	changes := &myentity.Transaction{
		// ID dan UserID jangan diubah di sini, tapi di GORM Update call akan difilter berdasarkan oldData
		Amount:          req.Amount,
		Type:            myentity.TransactionType(req.Type),
		TransactionDate: parsedDate,
		UpdatedAt:       helper.DatetimeNowJakarta(), // Menggunakan helper
		// Handle Description dan CategoryID menggunakan sql.NullXXX
		Description:     sql.NullString{String: *req.Description, Valid: req.Description != nil},
		CategoryID:      newCategoryID,
	}

	// Panggil repository untuk update (oldData digunakan GORM untuk WHERE, changes adalah nilai baru)
	err = u.TransactionRepo.Update(ctx, nil, oldData, changes) // oldData untuk menemukan record, changes untuk data yang diubah
	if err != nil {
		helper.LogError(funcName, "TransactionRepo.Update", err, logFields, "")
		return err
	}

	return nil
}

// Delete menghapus transaksi berdasarkan ID dan memastikan milik user yang benar.
func (u *CrudTransaction) Delete(ctx context.Context, id int64, userID int64) error {
	funcName := "CrudTransaction.Delete"
	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"id":      fmt.Sprintf("%d", id),
	}

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// Validasi apakah data dengan ID tersebut ada dan milik user yang benar
	// Menggunakan GetByIDAndUserID untuk memastikan otorisasi di lapisan usecase
	_, err := u.TransactionRepo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		helper.LogError(funcName, "GetByIDAndUserID", err, logFields, "Error getting transaction for delete (authorization check)")
		return err // Error akan berupa ErrRecordNotFound atau error lain dari repo
	}

	// Lakukan delete (repository sudah memfilter berdasarkan user_id)
	err = u.TransactionRepo.DeleteByIDAndUserID(ctx, nil, id, userID)
	if err != nil {
		helper.LogError(funcName, "TransactionRepo.DeleteByIDAndUserID", err, logFields, "")
		return err
	}

	return nil
}

// GetDailySummary mengambil ringkasan transaksi harian untuk user tertentu.
func (u *CrudTransaction) GetDailySummary(ctx context.Context, userID int64, startDate, endDate string) ([]map[string]interface{}, error) {
	funcName := "CrudTransaction.GetDailySummary"
	logFields := generalEntity.CaptureFields{
		"user_id":    strconv.FormatInt(userID, 10),
		"start_date": startDate,
		"end_date":   endDate,
	}

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return nil, apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// Validasi tanggal
	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		helper.LogError(funcName, "time.Parse", err, logFields, "Invalid start_date format")
		return nil, apperr.ErrInvalidRequest().SetDetail("Invalid start_date format. Use YYYY-MM-DD.")
	}
	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		helper.LogError(funcName, "time.Parse", err, logFields, "Invalid end_date format")
		return nil, apperr.ErrInvalidRequest().SetDetail("Invalid end_date format. Use YYYY-MM-DD.")
	}

	result, err := u.TransactionRepo.GetDailySummaryByUserID(ctx, userID, startDate, endDate)
	if err != nil {
		helper.LogError(funcName, "TransactionRepo.GetDailySummaryByUserID", err, logFields, "")
		return nil, err
	}

	return result, nil
}

// GetSummaryByCategoryAndType mengambil ringkasan transaksi per kategori dan tipe untuk user tertentu.
func (u *CrudTransaction) GetSummaryByCategoryAndType(ctx context.Context, userID int64, startDate, endDate string) ([]usecaseEntity.TransactionSummaryResponse, error) {
	funcName := "CrudTransaction.GetSummaryByCategoryAndType"
	logFields := generalEntity.CaptureFields{
		"user_id":    strconv.FormatInt(userID, 10),
		"start_date": startDate,
		"end_date":   endDate,
	}

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return nil, apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// Validasi tanggal
	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		helper.LogError(funcName, "time.Parse", err, logFields, "Invalid start_date format")
		return nil, apperr.ErrInvalidRequest().SetDetail("Invalid start_date format. Use YYYY-MM-DD.")
	}
	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		helper.LogError(funcName, "time.Parse", err, logFields, "Invalid end_date format")
		return nil, apperr.ErrInvalidRequest().SetDetail("Invalid end_date format. Use YYYY-MM-DD.")
	}

	// Panggil repository untuk mendapatkan data summary
	data, err := u.TransactionRepo.GetSummaryByCategoryAndTypeByUserID(ctx, userID, startDate, endDate)
	if err != nil {
		helper.LogError(funcName, "TransactionRepo.GetSummaryByCategoryAndTypeByUserID", err, logFields, "")
		return nil, err
	}

	// Map hasil dari repository ke DTO respons
	var result []usecaseEntity.TransactionSummaryResponse
	for _, row := range data {
		var categoryName *string
		if row.CategoryName.Valid {
			categoryName = &row.CategoryName.String
		}
		result = append(result, usecaseEntity.TransactionSummaryResponse{
			CategoryName: categoryName,
			Type:         usecaseEntity.TransactionTypeString(row.Type), // Konversi ke DTO type
			TotalAmount:  row.TotalAmount,
		})
	}

	return result, nil
}