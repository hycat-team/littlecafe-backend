// Package repositories holds GORM-backed building blocks shared by every
// service's persistence layer.
package repositories

import (
	"context"

	"github.com/hycat-team/littlecafe-backend/shared/domain/repositories"
	"github.com/hycat-team/littlecafe-backend/shared/infrastructure/db"

	"gorm.io/gorm"
)

// GenericRepository is a GORM implementation of IGenericRepository.
//
// It is transaction-aware: every method resolves the active transaction from
// the context (bound by IUnitOfWork.Execute via db.InjectTx) before falling
// back to the base connection. Preloads are fixed at construction so inherited
// lookups and custom queries load the same relations.
type GenericRepository[T any, ID any] struct {
	baseDB   *gorm.DB
	preloads []string
}

// NewGenericRepository builds a repository for entity T. Relations passed as
// preloads are eager-loaded by every inherited method.
func NewGenericRepository[T any, ID any](baseDB *gorm.DB, preloads ...string) *GenericRepository[T, ID] {
	return &GenericRepository[T, ID]{baseDB: baseDB, preloads: preloads}
}

// GetDB returns the transaction bound to ctx when present, otherwise the base
// connection. Custom query methods MUST use it so they join an open unit of
// work.
func (r *GenericRepository[T, ID]) GetDB(ctx context.Context) *gorm.DB {
	if tx := db.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return r.baseDB.WithContext(ctx)
}

// GetQueryWithPreload returns a query with the configured preloads applied.
func (r *GenericRepository[T, ID]) GetQueryWithPreload(ctx context.Context) *gorm.DB {
	query := r.GetDB(ctx)
	for _, preload := range r.preloads {
		query = query.Preload(preload)
	}
	return query
}

func (r *GenericRepository[T, ID]) GetByID(ctx context.Context, id ID) (*T, error) {
	var entity T
	if err := r.GetQueryWithPreload(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericRepository[T, ID]) GetByIDs(ctx context.Context, ids []ID) ([]*T, error) {
	if len(ids) == 0 {
		return []*T{}, nil
	}
	var entities []*T
	if err := r.GetQueryWithPreload(ctx).Where("id IN ?", ids).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GenericRepository[T, ID]) GetAll(ctx context.Context) ([]*T, error) {
	entities := make([]*T, 0)
	if err := r.GetQueryWithPreload(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GenericRepository[T, ID]) Create(ctx context.Context, entity *T) error {
	return r.GetDB(ctx).Create(entity).Error
}

func (r *GenericRepository[T, ID]) Update(ctx context.Context, entity *T) error {
	return r.GetDB(ctx).Save(entity).Error
}

func (r *GenericRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	var entity T
	return r.GetDB(ctx).Delete(&entity, "id = ?", id).Error
}

func (r *GenericRepository[T, ID]) BatchUpdate(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	return r.GetDB(ctx).Save(entities).Error
}

func (r *GenericRepository[T, ID]) BatchUpdateMap(ctx context.Context, ids []ID, values map[string]any) error {
	if len(ids) == 0 || len(values) == 0 {
		return nil
	}
	var entity T
	return r.GetDB(ctx).Model(&entity).Where("id IN ?", ids).Updates(values).Error
}

var _ repositories.IGenericRepository[any, any] = (*GenericRepository[any, any])(nil)
