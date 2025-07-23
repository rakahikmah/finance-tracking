package handler

import (
	"net/http"
	"strconv" // Untuk mengkonversi string ke int64

	fiber "github.com/gofiber/fiber/v2"
	"github.com/rakahikmah/finance-tracking/internal/http/middleware"
	"github.com/rakahikmah/finance-tracking/internal/parser"
	"github.com/rakahikmah/finance-tracking/internal/presenter/json"
	transactions_usecase "github.com/rakahikmah/finance-tracking/internal/usecase/transactions" // Import usecase Transactions Anda
	usecaseEntity "github.com/rakahikmah/finance-tracking/internal/usecase/transactions/entity" // Import DTO usecase Transactions Anda

	apperr "github.com/rakahikmah/finance-tracking/error"
)

// TransactionHandler adalah handler HTTP untuk operasi Transaction.
type TransactionHandler struct {
	parser            parser.Parser
	presenter         json.JsonPresenter
	CrudTransactionUsecase transactions_usecase.ICrudTransaction // Menggunakan interface usecase Transaction
}

// NewTransactionHandler adalah konstruktor untuk TransactionHandler.
func NewTransactionHandler(
	parser parser.Parser,
	presenter json.JsonPresenter,
	CrudTransactionUsecase transactions_usecase.ICrudTransaction,
) *TransactionHandler {
	return &TransactionHandler{parser, presenter, CrudTransactionUsecase}
}

// Register mendaftarkan rute-rute API untuk Transaction.
func (h *TransactionHandler) Register(app fiber.Router) {
	// Semua rute ini akan memerlukan otentikasi JWT
	app.Post("/transactions", middleware.VerifyJWTToken, h.Create)
	app.Get("/transactions", middleware.VerifyJWTToken, h.GetAll)
	app.Get("/transactions/summary", middleware.VerifyJWTToken, h.GetDailySummary) // Rute baru untuk summary
	app.Put("/transactions/:id", middleware.VerifyJWTToken, h.Update)
	app.Get("/transactions/summary-by-category-type", middleware.VerifyJWTToken, h.GetSummaryByCategoryAndType)
	app.Delete("/transactions/:id", middleware.VerifyJWTToken, h.Delete)
}

// Create menangani permintaan POST untuk membuat transaksi baru.
func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	var req usecaseEntity.TransactionReq // Menggunakan TransactionReq dari usecase entity

	// ParserBodyRequestWithUserID akan meng-unmarshal body request
	// DAN mengambil userID dari Fiber context (yang sudah ditaruh oleh JWT middleware)
	// lalu menyisipkannya ke req.UserID (field exported).
	err := h.parser.ParserBodyRequestWithUserID(c, &req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	// Ambil userID dari Fiber context. Ini adalah userID yang terautentikasi.
	// Ini krusial karena kita akan meneruskannya ke usecase.
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context (from JWT)."))
	}

	// Memanggil usecase.Create dengan userID sebagai parameter terpisah
	err = h.CrudTransactionUsecase.Create(c.Context(), userID, req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, nil, "Transaction created successfully", http.StatusCreated)
}

// GetAll menangani permintaan GET untuk mendapatkan semua transaksi user.
func (h *TransactionHandler) GetAll(c *fiber.Ctx) error {
	// Ambil userID dari Fiber context
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context (from JWT)."))
	}

	// Memanggil usecase.GetAll dengan userID
	result, err := h.CrudTransactionUsecase.GetAll(c.Context(), userID)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, result, "Transactions retrieved successfully", http.StatusOK)
}

// GetDailySummary menangani permintaan GET untuk ringkasan transaksi harian.
func (h *TransactionHandler) GetDailySummary(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context (from JWT)."))
	}

	// Ambil parameter tanggal dari query string (misal: /summary?start_date=2023-01-01&end_date=2023-01-31)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validasi dasar parameter tanggal
	if startDate == "" || endDate == "" {
		return h.presenter.BuildError(c, apperr.ErrInvalidRequest().SetDetail("start_date and end_date query parameters are required for summary."))
	}

	result, err := h.CrudTransactionUsecase.GetDailySummary(c.Context(), userID, startDate, endDate)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, result, "Daily transaction summary retrieved successfully", http.StatusOK)
}


// Update menangani permintaan PUT untuk memperbarui transaksi.
func (h *TransactionHandler) Update(c *fiber.Ctx) error {
	// Ambil ID transaksi dari parameter URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return h.presenter.BuildError(c, apperr.ErrInvalidRequest().SetDetail("Invalid transaction ID format."))
	}

	var req usecaseEntity.TransactionReq
	// Gunakan ParserBodyRequestWithUserID agar userID otomatis disisipkan ke req.UserID
	err = h.parser.ParserBodyRequestWithUserID(c, &req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	// Ambil userID dari Fiber context (penting untuk otorisasi di usecase)
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context (from JWT)."))
	}

	// Memanggil usecase.Update dengan ID transaksi dan userID
	err = h.CrudTransactionUsecase.Update(c.Context(), id, userID, req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, nil, "Transaction updated successfully", http.StatusOK)
}

// Delete menangani permintaan DELETE untuk menghapus transaksi.
func (h *TransactionHandler) Delete(c *fiber.Ctx) error {
	// Ambil ID transaksi dari parameter URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return h.presenter.BuildError(c, apperr.ErrInvalidRequest().SetDetail("Invalid transaction ID format."))
	}

	// Ambil userID dari Fiber context
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context (from JWT)."))
	}

	// Memanggil usecase.Delete dengan ID transaksi dan userID
	err = h.CrudTransactionUsecase.Delete(c.Context(), id, userID)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, nil, "Transaction deleted successfully", http.StatusOK)
}


// GetSummaryByCategoryAndType menangani permintaan GET untuk ringkasan transaksi per kategori dan tipe.
func (h *TransactionHandler) GetSummaryByCategoryAndType(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context (from JWT)."))
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return h.presenter.BuildError(c, apperr.ErrInvalidRequest().SetDetail("start_date and end_date query parameters are required for summary."))
	}

	result, err := h.CrudTransactionUsecase.GetSummaryByCategoryAndType(c.Context(), userID, startDate, endDate)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, result, "Transaction summary by category and type retrieved successfully", http.StatusOK)
}