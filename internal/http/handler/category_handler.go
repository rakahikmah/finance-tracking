package handler

import (
	"net/http"
	"strconv" // Untuk mengkonversi string ke int64

	fiber "github.com/gofiber/fiber/v2"
	"github.com/rakahikmah/finance-tracking/internal/http/middleware"
	"github.com/rakahikmah/finance-tracking/internal/parser"
	"github.com/rakahikmah/finance-tracking/internal/presenter/json"
	category_usecase "github.com/rakahikmah/finance-tracking/internal/usecase/category"     // Import usecase Category Anda
	usecaseEntity "github.com/rakahikmah/finance-tracking/internal/usecase/category/entity" // Import DTO usecase Category Anda

	apperr "github.com/rakahikmah/finance-tracking/error"
)

// CategoryHandler adalah handler HTTP untuk operasi Category.
type CategoryHandler struct {
	parser              parser.Parser
	presenter           json.JsonPresenter
	CrudCategoryUsecase category_usecase.ICrudCategory // Menggunakan interface usecase Category
}

// NewCategoryHandler adalah konstruktor untuk CategoryHandler.
func NewCategoryHandler(
	parser parser.Parser,
	presenter json.JsonPresenter,
	CrudCategoryUsecase category_usecase.ICrudCategory,
) *CategoryHandler {
	return &CategoryHandler{parser, presenter, CrudCategoryUsecase}
}

// Register mendaftarkan rute-rute API untuk Category.
func (h *CategoryHandler) Register(app fiber.Router) {
	// Semua rute ini akan memerlukan otentikasi JWT
	app.Post("/categories", middleware.VerifyJWTToken, h.Create)
	app.Get("/categories", middleware.VerifyJWTToken, h.GetAll)
	app.Put("/categories/:id", middleware.VerifyJWTToken, h.Update)    // Tambahkan middleware JWT untuk Update
	app.Delete("/categories/:id", middleware.VerifyJWTToken, h.Delete) // Tambahkan middleware JWT untuk Delete
}

// Create menangani permintaan POST untuk membuat kategori baru.
func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var req usecaseEntity.CategoryReq // Menggunakan CategoryReq dari usecase entity

	err := h.parser.ParserBodyRequestWithUserID(c, &req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}


	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context 123."))
	}

	// Memanggil usecase.Create dengan userID sebagai parameter terpisah
	err = h.CrudCategoryUsecase.Create(c.Context(), userID, req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, nil, "Category created successfully", http.StatusCreated)
}

// GetAll menangani permintaan GET untuk mendapatkan semua kategori user.
func (h *CategoryHandler) GetAll(c *fiber.Ctx) error {
	// Ambil userID dari Fiber context
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context."))
	}

	// Memanggil usecase.GetAll dengan userID
	result, err := h.CrudCategoryUsecase.GetAll(c.Context(), userID)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, result, "Categories retrieved successfully", http.StatusOK)
}

// Update menangani permintaan PUT untuk memperbarui kategori.
func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	// Ambil ID kategori dari parameter URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64) // Gunakan strconv.ParseInt untuk lebih robust
	if err != nil || id <= 0 {
		return h.presenter.BuildError(c, apperr.ErrInvalidRequest().SetDetail("Invalid category ID format."))
	}

	var req usecaseEntity.CategoryReq
	// Gunakan ParserBodyRequestWithUserID agar userID otomatis disisipkan ke req.userID
	err = h.parser.ParserBodyRequestWithUserID(c, &req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	// Ambil userID dari Fiber context
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context."))
	}

	// Memanggil usecase.Update dengan ID kategori dan userID
	err = h.CrudCategoryUsecase.Update(c.Context(), id, userID, req)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, nil, "Category updated successfully", http.StatusOK)
}

// Delete menangani permintaan DELETE untuk menghapus kategori.
func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	// Ambil ID kategori dari parameter URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return h.presenter.BuildError(c, apperr.ErrInvalidRequest().SetDetail("Invalid category ID format."))
	}

	// Ambil userID dari Fiber context
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID == 0 {
		return h.presenter.BuildError(c, apperr.ErrUnauthorized().SetDetail("User ID not found in context."))
	}

	// Memanggil usecase.Delete dengan ID kategori dan userID
	err = h.CrudCategoryUsecase.Delete(c.Context(), id, userID)
	if err != nil {
		return h.presenter.BuildError(c, err)
	}

	return h.presenter.BuildSuccess(c, nil, "Category deleted successfully", http.StatusOK)
}
