package repositories

import "gorm.io/gorm"

// ApplyAdminPagination applies offset pagination for administrative tables that
// need random access. Client-facing endpoints MUST use cursor pagination
// instead (see application/dto.CursorPage).
func ApplyAdminPagination(db *gorm.DB, page, limit int) *gorm.DB {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return db.Offset((page - 1) * limit).Limit(limit)
}
