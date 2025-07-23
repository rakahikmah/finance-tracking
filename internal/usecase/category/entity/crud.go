package entity


type CategoryReq struct {
	Name      string `json:"name" validate:"required" name:"Nama Kategori"`
	userID int64  `validate:"required" name:"ID Pembuat"`
}

type CategoryResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedBy int64  `json:"created_by"`
	CreatedAt string `json:"created_at"` // Biasanya diubah ke string untuk format JSON
	UpdatedAt string `json:"updated_at"` // Biasanya diubah ke string untuk format JSON
}


func (r *CategoryReq) SetUserID(userID int64) {
	r.userID = userID
}

