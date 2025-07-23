package error

import (
	"fmt" // Import fmt untuk Sprintf di method Error()
	"net/http"

	"github.com/rakahikmah/finance-tracking/entity" // Pastikan import path ini benar
)

// CustomErrorResponse merepresentasikan struktur error kustom untuk API.
type CustomErrorResponse struct {
	Message  string `json:"message,omitempty"`
	ErrCode  string `json:"code,omitempty"`
	HTTPCode int    `json:"http_code"`
	Detail   string `json:"detail,omitempty"` // <-- Field baru untuk detail tambahan
}

// CustomErrorResponseWithMeta adalah struktur error dengan metadata tambahan.
type CustomErrorResponseWithMeta struct {
	Message  string               `json:"message,omitempty"`
	ErrCode  string               `json:"code,omitempty"`
	HTTPCode int                  `json:"http_code"`
	Meta     []entity.ErrorResponse `json:"meta,omitempty"`
}

// SetDetail adalah method untuk menambahkan detail ke CustomErrorResponse.
// Method ini mengembalikan CustomErrorResponse sehingga bisa di-chaining.
func (c CustomErrorResponse) SetDetail(detail string) CustomErrorResponse {
	c.Detail = detail
	return c
}

// Error adalah method untuk memenuhi interface error Go.
// Ini mengembalikan representasi string dari error.
func (c CustomErrorResponse) Error() string {
	if c.Detail != "" {
		return fmt.Sprintf("%s: %s", c.Message, c.Detail)
	}
	return c.Message
}

// --- Fungsi Pembuat Error Umum ---

func ErrRecordNotFound() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.DATA_NOT_FOUND_MSG,
		ErrCode:  entity.BAD_REQUEST_MSG, // Anda mungkin ingin kode error yang lebih spesifik di sini, misalnya "E404"
		HTTPCode: http.StatusNotFound,
	}
}

func ErrUserNotFound() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.USER_NOT_FOUND_MSG,
		ErrCode:  entity.BAD_REQUEST_MSG, // Atau kode yang lebih spesifik
		HTTPCode: http.StatusNotFound,
	}
}

func ErrInvalidEmailOrPassword() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.INVALID_AUTH_MSG,
		ErrCode:  entity.INVALID_AUTH_CODE,
		HTTPCode: http.StatusUnauthorized,
	}
}

func ErrInvalidToken() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.INVALID_TOKEN_MSG,
		ErrCode:  entity.INVALID_TOKEN_CODE,
		HTTPCode: http.StatusUnauthorized,
	}
}

func ErrInvalidPayload(meta []entity.ErrorResponse) CustomErrorResponseWithMeta {
	return CustomErrorResponseWithMeta{
		Message:  entity.INVALID_PAYLOAD_MSG,
		ErrCode:  entity.INVALID_PAYLOAD_CODE,
		HTTPCode: http.StatusUnprocessableEntity,
		Meta:     meta,
	}
}

func ErrGeneralInvalid() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.GENERAL_ERROR_MESSAGE,
		ErrCode:  entity.BAD_REQUEST_MSG,
		HTTPCode: http.StatusUnprocessableEntity,
	}
}

func ErrInvalidRequest() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.INVALID_PAYLOAD_MSG, // Umumnya invalid request = invalid payload
		ErrCode:  entity.BAD_REQUEST_MSG,
		HTTPCode: http.StatusUnprocessableEntity, // Atau HttpStatusBadRequest
	}
}

// ErrUnauthorized mengembalikan CustomErrorResponse untuk akses tidak sah.
func ErrUnauthorized() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.UNAUTHORIZED_MSG, // Pastikan ini didefinisikan di entity
		ErrCode:  entity.UNAUTHORIZED_CODE, // Pastikan ini didefinisikan di entity
		HTTPCode: http.StatusUnauthorized,
	}
}

// ErrConflict mengembalikan CustomErrorResponse untuk konflik data (misalnya, duplikasi).
func ErrConflict() CustomErrorResponse {
	return CustomErrorResponse{
		Message:  entity.CONFLICT_MSG, // <-- Harus didefinisikan di entity
		ErrCode:  entity.CONFLICT_CODE, // <-- Harus didefinisikan di entity
		HTTPCode: http.StatusConflict,
	}
}

func CustomError(message string, errCode string, httpCode int) CustomErrorResponse {
	return CustomErrorResponse{
		Message:  message,
		ErrCode:  errCode,
		HTTPCode: httpCode,
	}
}