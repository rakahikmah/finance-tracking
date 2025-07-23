package mysql

import (
	"context"

	"github.com/rakahikmah/finance-tracking/config" // Sesuaikan import path projectmu
	"github.com/rakahikmah/finance-tracking/internal/helper" // Sesuaikan import path projectmu
	"github.com/rakahikmah/finance-tracking/internal/repository/mysql/entity" // Menggunakan entity.Category yang sudah kita buat

	apperr "github.com/rakahikmah/finance-tracking/error" // Sesuaikan import path projectmu

	errwrap "github.com/pkg/errors"
	"gorm.io/gorm"
)

// ICategoryRepository mendefinisikan interface untuk operasi CRUD pada entitas Category.
type ICategoryRepository interface {
	TrxSupportRepo // Warisan dari interface transaksi (biasanya ada di file mysql/common.go)
	GetByID(ctx context.Context, ID int64) (e *entity.Category, err error)
	Create(ctx context.Context, dbTrx TrxObj, params *entity.Category, nonZeroVal bool) error
	Update(ctx context.Context, dbTrx TrxObj, params *entity.Category, changes *entity.Category) (err error)
	DeleteByID(ctx context.Context, dbTrx TrxObj, id int64) error
	GetAll(ctx context.Context, userID int64) (result []*entity.Category, err error) // Menambahkan userID untuk filter
	GetByUserIDAndName(ctx context.Context, userID int64, name string) (e *entity.Category, err error) // Tambahan untuk cek duplikasi nama per user
}

// CategoryRepository adalah implementasi repository untuk entitas Category.
type CategoryRepository struct {
	GormTrxSupport // Warisan dari struct untuk dukungan transaksi
}

// NewCategoryRepository membuat instance baru dari CategoryRepository.
func NewCategoryRepository(mysql *config.Mysql) *CategoryRepository {
	return &CategoryRepository{GormTrxSupport{db: mysql.DB}}
}

// GetAll mengambil semua kategori yang dimiliki oleh user tertentu.
func (r *CategoryRepository) GetAll(ctx context.Context, userID int64) (result []*entity.Category, err error) {
	funcName := "CategoryRepository.GetAll"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	// Menambahkan filter WHERE created_by = ?
	err = r.db.Where("created_by = ?", userID).Find(&result).Error
	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		// Jika tidak ada record, kembalikan slice kosong, bukan error
		return []*entity.Category{}, nil 
	}
    if err != nil {
        return nil, errwrap.Wrap(err, funcName)
    }

	return result, nil
}

// GetByID mengambil kategori berdasarkan ID.
// Kategori juga harus dimiliki oleh user tertentu untuk alasan keamanan.
func (r *CategoryRepository) GetByID(ctx context.Context, ID int64) (result *entity.Category, err error) {
	funcName := "CategoryRepository.GetByID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	// Menggunakan GORM Find atau First lebih idiomatik daripada Raw SQL
	err = r.db.First(&result, ID).Error // Find by primary key ID
	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrRecordNotFound()
	}
    if err != nil {
        return nil, errwrap.Wrap(err, funcName)
    }

	return result, nil
}

// GetByUserIDAndName mengambil kategori berdasarkan user ID dan nama.
// Berguna untuk memeriksa duplikasi nama kategori per user.
func (r *CategoryRepository) GetByUserIDAndName(ctx context.Context, userID int64, name string) (result *entity.Category, err error) {
	funcName := "CategoryRepository.GetByUserIDAndName"

	if err := helper.CheckDeadline(ctx); err != nil {
		return nil, errwrap.Wrap(err, funcName)
	}

	err = r.db.Where("created_by = ? AND name = ?", userID, name).First(&result).Error
	if errwrap.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrRecordNotFound()
	}
    if err != nil {
        return nil, errwrap.Wrap(err, funcName)
    }
	return result, nil
}


// Create membuat kategori baru.
func (r *CategoryRepository) Create(ctx context.Context, dbTrx TrxObj, params *entity.Category, nonZeroVal bool) error {
	funcName := "CategoryRepository.Create"

	if err := helper.CheckDeadline(ctx); err != nil {
		return errwrap.Wrap(err, funcName)
	}

	// helper.NonZeroCols akan memilih kolom yang tidak nol atau kosong untuk dimasukkan.
	cols := helper.NonZeroCols(params, nonZeroVal)
	return r.Trx(dbTrx).Select(cols).Create(&params).Error
}

// Update memperbarui kategori yang ada.
func (r *CategoryRepository) Update(ctx context.Context, dbTrx TrxObj, params *entity.Category, changes *entity.Category) error {
	funcName := "CategoryRepository.Update"

	if err := helper.CheckDeadline(ctx); err != nil {
		return errwrap.Wrap(err, funcName)
	}

	if params.ID == 0 {
		return errwrap.Wrap(apperr.ErrInvalidRequest(), funcName)
	}

	// Model(params) akan menggunakan ID dari params untuk mencari record.
	db := r.Trx(dbTrx).Model(params)

	var err error
	if changes != nil {
		// Updates(*changes) hanya akan mengupdate kolom yang non-zero di struct changes.
		err = db.Updates(*changes).Error
	} else {
		// helper.StructToMap akan mengkonversi struct params ke map, lalu Updates akan memperbarui semua kolom di map.
		err = db.Updates(helper.StructToMap(params, false)).Error
	}

	if err != nil {
		return errwrap.Wrap(err, funcName)
	}

	return nil
}

// DeleteByID menghapus kategori berdasarkan ID.
func (r *CategoryRepository) DeleteByID(ctx context.Context, dbTrx TrxObj, id int64) error {
	funcName := "CategoryRepository.DeleteByID"

	if err := helper.CheckDeadline(ctx); err != nil {
		return errwrap.Wrap(err, funcName)
	}

	// Delete(&entity.Category{}) akan menghapus record dari tabel "categories" dengan ID yang sesuai.
	err := r.Trx(dbTrx).Where("id = ?", id).Delete(&entity.Category{}).Error
	if err != nil {
		return errwrap.Wrap(err, funcName) // Menggunakan errwrap.Wrap untuk konsistensi
	}

	return nil
}