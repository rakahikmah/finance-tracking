package category_usecase // Nama paket harus berbeda dari 'entity'

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	generalEntity "github.com/rakahikmah/finance-tracking/entity"
	"github.com/rakahikmah/finance-tracking/internal/helper"
	"github.com/rakahikmah/finance-tracking/internal/repository/mysql"
	myentity "github.com/rakahikmah/finance-tracking/internal/repository/mysql/entity"
	"github.com/rakahikmah/finance-tracking/internal/usecase/category/entity"

	apperr "github.com/rakahikmah/finance-tracking/error"
)

// CrudCategory adalah struct yang akan menampung dependensi repository.
type CrudCategory struct {
	CategoryRepo mysql.ICategoryRepository
}

// NewCrudCategory adalah konstruktor untuk CrudCategory.
func NewCrudCategory(
	CategoryRepo mysql.ICategoryRepository,
) *CrudCategory {
	return &CrudCategory{CategoryRepo: CategoryRepo}
}

// ICrudCategory mendefinisikan interface untuk operasi CRUD pada Category.
type ICrudCategory interface {
	// Ini sudah benar
	Create(ctx context.Context, userID int64, req entity.CategoryReq) error
	GetAll(ctx context.Context, userID int64) ([]entity.CategoryResponse, error)
	Update(ctx context.Context, id int64, userID int64, req entity.CategoryReq) error
	Delete(ctx context.Context, id int64, userID int64) error
}

func (u *CrudCategory) Create(ctx context.Context, userID int64, req entity.CategoryReq) error {
	funcName := "CrudCategory.Create"

	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, nil, "UserID tidak ditemukan")
		return apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10), // Sekarang `userID` di sini merujuk ke parameter
		"name":    req.Name,
	}

	// 1. Cek duplikasi nama kategori untuk user yang sama
	existingCategory, err := u.CategoryRepo.GetByUserIDAndName(ctx, userID, req.Name) // Menggunakan parameter `userID`
	if err != nil && !errors.Is(err, apperr.ErrRecordNotFound()) {
		helper.LogError(funcName, "GetByUserIDAndName", err, logFields, "Error checking for existing category name")
		return err
	}
	if existingCategory != nil {
		helper.LogError(funcName, "GetByUserIDAndName", errors.New("category name already exists for this user"), logFields, "")
		return apperr.ErrConflict().SetDetail(fmt.Sprintf("Category with name '%s' already exists for this user.", req.Name))
	}

	// 2. Siapkan data untuk disimpan ke database
	data := &myentity.Category{
		Name:      req.Name,
		CreatedAt: helper.DatetimeNowJakarta(),
		UpdatedAt: helper.DatetimeNowJakarta(),
		CreatedBy: userID, // Menggunakan parameter `userID`
	}

	// 3. Panggil repository untuk membuat record
	err = u.CategoryRepo.Create(ctx, nil, data, false)
	if err != nil {
		helper.LogError(funcName, "CategoryRepo.Create", err, logFields, "")
		return err
	}

	return nil
}

// // GetAll mengambil semua kategori untuk user tertentu.
func (u *CrudCategory) GetAll(ctx context.Context, userID int64) ([]entity.CategoryResponse, error) {
	funcName := "CrudCategory.GetAll"
	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"layer":   "usecase",
	}

	// Pastikan UserID valid
	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return nil, apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// Ambil data dari repository, dengan filter userID
	data, err := u.CategoryRepo.GetAll(ctx, userID)
	if err != nil {
		helper.LogError(funcName, "CategoryRepo.GetAll", err, logFields, "")
		return nil, err
	}

	// Mapping ke response DTO
	var result []entity.CategoryResponse
	for _, row := range data {
		result = append(result, entity.CategoryResponse{
			ID:        row.ID,
			Name:      row.Name,
			CreatedBy: row.CreatedBy,
			CreatedAt: helper.ConvertToJakartaTime(row.CreatedAt), // Konversi time.Time ke string
			UpdatedAt: helper.ConvertToJakartaTime(row.UpdatedAt), // Konversi time.Time ke string
		})
	}

	return result, nil
}

// // Update memperbarui kategori berdasarkan ID dan memastikan milik user yang benar.
func (u *CrudCategory) Update(ctx context.Context, id int64, userID int64, req entity.CategoryReq) error {
	funcName := "CrudCategory.Update"
	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"id":      fmt.Sprintf("%d", id),
	}

	// Validasi UserID
	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// 1. Ambil data lama dari database
	oldData, err := u.CategoryRepo.GetByID(ctx, id)
	if err != nil {
		helper.LogError(funcName, "GetByID", err, logFields, "Error getting existing category")
		return err
	}

	// 2. Otorisasi: Pastikan kategori yang akan diupdate adalah milik user yang sedang login
	if oldData.CreatedBy != userID {
		helper.LogError(funcName, "Authorization", errors.New("unauthorized access to category"), logFields, "User tried to update category not owned by them")
		return apperr.ErrUnauthorized().SetDetail("You are not authorized to update this category.")
	}

	// 3. (Opsional) Cek duplikasi nama jika nama diubah
	if oldData.Name != req.Name { // Jika nama kategori diubah
		existingCategory, err := u.CategoryRepo.GetByUserIDAndName(ctx, userID, req.Name)
		if err != nil && !errors.Is(err, apperr.ErrRecordNotFound()) {
			helper.LogError(funcName, "GetByUserIDAndName", err, logFields, "Error checking for existing category name on update")
			return err
		}
		if existingCategory != nil && existingCategory.ID != id { // Jika nama baru sudah ada di kategori lain milik user ini
			helper.LogError(funcName, "GetByUserIDAndName", errors.New("category name already exists for this user"), logFields, "")
			return apperr.ErrConflict().SetDetail(fmt.Sprintf("Category with name '%s' already exists for this user.", req.Name))
		}
	}

	// 4. Siapkan perubahan data
	changes := &myentity.Category{
		Name:      req.Name,
		UpdatedAt: helper.DatetimeNowJakarta(), // Update UpdatedAt
	}

	// 5. Panggil repository untuk update
	err = u.CategoryRepo.Update(ctx, nil, oldData, changes)
	if err != nil {
		helper.LogError(funcName, "CategoryRepo.Update", err, logFields, "")
		return err
	}

	return nil
}

// Delete menghapus kategori berdasarkan ID dan memastikan milik user yang benar.
func (u *CrudCategory) Delete(ctx context.Context, id int64, userID int64) error {
	funcName := "CrudCategory.Delete"
	logFields := generalEntity.CaptureFields{
		"user_id": strconv.FormatInt(userID, 10),
		"id":      fmt.Sprintf("%d", id),
	}

	// Validasi UserID
	if userID == 0 {
		err := errors.New("user ID tidak ditemukan di konteks request")
		helper.LogError(funcName, "validasi request", err, logFields, "UserID tidak ditemukan")
		return apperr.ErrInvalidRequest().SetDetail("User ID is required")
	}

	// 1. Validasi apakah data dengan ID tersebut ada dan milik user yang benar
	oldData, err := u.CategoryRepo.GetByID(ctx, id)
	if err != nil {
		helper.LogError(funcName, "GetByID", err, logFields, "Error getting category for delete")
		return err
	}

	// 2. Otorisasi: Pastikan kategori yang akan dihapus adalah milik user yang sedang login
	if oldData.CreatedBy != userID {
		helper.LogError(funcName, "Authorization", errors.New("unauthorized access to category"), logFields, "User tried to delete category not owned by them")
		return apperr.ErrUnauthorized().SetDetail("You are not authorized to delete this category.")
	}

	// 3. Lakukan delete
	err = u.CategoryRepo.DeleteByID(ctx, nil, id)
	if err != nil {
		helper.LogError(funcName, "CategoryRepo.DeleteByID", err, logFields, "")
		return err
	}

	return nil
}
