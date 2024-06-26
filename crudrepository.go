package repository_gorm

import (
	"context"
	"github.com/ace-zhaoy/errors"
	"github.com/ace-zhaoy/go-repository"
	"github.com/ace-zhaoy/go-repository/contract"
	"github.com/ace-zhaoy/go-utils/uslice"
	"gorm.io/gorm"
	"strings"
)

type CrudRepository[ID comparable, ENTITY contract.ENTITY[ID]] struct {
	db                *gorm.DB
	unscoped          bool
	idField           string
	softDeleteField   string
	softDeleteEnabled bool
}

var _ contract.CrudRepository[int64, contract.ENTITY[int64]] = (*CrudRepository[int64, contract.ENTITY[int64]])(nil)

func NewCrudRepository[ID comparable, ENTITY contract.ENTITY[ID]](db *gorm.DB) *CrudRepository[ID, ENTITY] {
	var entity ENTITY
	softDeleteField := getDeletedAtField(entity)
	return &CrudRepository[ID, ENTITY]{
		db:                db,
		idField:           getIDField(entity),
		softDeleteField:   softDeleteField,
		softDeleteEnabled: softDeleteField != "",
	}
}

func (c *CrudRepository[ID, ENTITY]) clone() *CrudRepository[ID, ENTITY] {
	return &CrudRepository[ID, ENTITY]{
		db:                c.db,
		unscoped:          c.unscoped,
		idField:           c.idField,
		softDeleteField:   c.softDeleteField,
		softDeleteEnabled: c.softDeleteEnabled,
	}
}

func (c *CrudRepository[ID, ENTITY]) query() *gorm.DB {
	if c.unscoped {
		return c.db.Unscoped()
	}
	return c.db
}

func (c *CrudRepository[ID, ENTITY]) Unscoped() contract.CrudRepository[ID, ENTITY] {
	cc := c.clone()
	cc.unscoped = true
	return cc
}

func (c *CrudRepository[ID, ENTITY]) IDField() string {
	return c.idField
}

func (c *CrudRepository[ID, ENTITY]) SoftDeleteField() string {
	return c.softDeleteField
}

func (c *CrudRepository[ID, ENTITY]) SoftDeleteEnabled() bool {
	return c.softDeleteEnabled
}

func (c *CrudRepository[ID, ENTITY]) Create(ctx context.Context, entity ENTITY) (id ID, err error) {
	defer errors.Recover(func(e error) { err = e })
	err = c.query().WithContext(ctx).Create(&entity).Error
	errors.Check(errors.WithStack(err))
	id = entity.GetID()
	return
}

func (c *CrudRepository[ID, ENTITY]) FindByID(ctx context.Context, id ID) (entity ENTITY, err error) {
	defer errors.Recover(func(e error) { err = errors.Wrap(e, "param: %v", id) })
	err = c.query().WithContext(ctx).First(&entity, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = repository.ErrNotFound.WrapStack(err)
	}
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) FindByIDs(ctx context.Context, ids []ID) (collection contract.Collection[ID, ENTITY], err error) {
	defer errors.Recover(func(e error) { err = errors.Wrap(e, "param: %v", ids) })
	var entities []ENTITY
	if len(ids) == 0 {
		collection = repository.NewCollection[ID, ENTITY](entities)
		return
	}
	err = c.query().WithContext(ctx).Find(&entities, ids).Error
	errors.Check(errors.WithStack(err))
	collection = repository.NewCollection[ID, ENTITY](entities)
	return
}

func (c *CrudRepository[ID, ENTITY]) FindByPage(ctx context.Context, limit, offset int, orders ...contract.Order) (collection contract.Collection[ID, ENTITY], err error) {
	defer errors.Recover(func(e error) { err = e })
	orderStr := strings.Join(uslice.Map(orders, func(order contract.Order) string { return order.ToString() }), ",")

	var entities []ENTITY
	err = c.query().WithContext(ctx).Limit(limit).Offset(offset).Order(orderStr).Find(&entities).Error
	errors.Check(errors.WithStack(err))
	collection = repository.NewCollection[ID, ENTITY](entities)
	return
}

func (c *CrudRepository[ID, ENTITY]) FindByFilter(ctx context.Context, filter map[string]any) (collection contract.Collection[ID, ENTITY], err error) {
	defer errors.Recover(func(e error) { err = e })
	var entities []ENTITY
	err = c.query().WithContext(ctx).Where(filter).Find(&entities).Error
	errors.Check(errors.WithStack(err))
	collection = repository.NewCollection[ID, ENTITY](entities)
	return
}

func (c *CrudRepository[ID, ENTITY]) FindByFilterWithPage(ctx context.Context, filter map[string]any, limit, offset int, orders ...contract.Order) (collection contract.Collection[ID, ENTITY], err error) {
	defer errors.Recover(func(e error) { err = e })
	orderStr := strings.Join(uslice.Map(orders, func(order contract.Order) string { return order.ToString() }), ",")

	var entities []ENTITY
	err = c.query().WithContext(ctx).Where(filter).Limit(limit).Offset(offset).Order(orderStr).Find(&entities).Error
	errors.Check(errors.WithStack(err))
	collection = repository.NewCollection[ID, ENTITY](entities)
	return
}

func (c *CrudRepository[ID, ENTITY]) FindAll(ctx context.Context) (collection contract.Collection[ID, ENTITY], err error) {
	defer errors.Recover(func(e error) { err = e })
	var entities []ENTITY
	err = c.query().WithContext(ctx).Find(&entities).Error
	errors.Check(errors.WithStack(err))
	collection = repository.NewCollection[ID, ENTITY](entities)
	return
}

func (c *CrudRepository[ID, ENTITY]) Count(ctx context.Context) (count int, err error) {
	defer errors.Recover(func(e error) { err = e })
	var cnt int64
	err = c.query().WithContext(ctx).Count(&cnt).Error
	errors.Check(errors.WithStack(err))
	count = int(cnt)
	return
}

func (c *CrudRepository[ID, ENTITY]) CountByFilter(ctx context.Context, filter map[string]any) (count int, err error) {
	defer errors.Recover(func(e error) { err = e })
	var cnt int64
	err = c.query().WithContext(ctx).Where(filter).Count(&cnt).Error
	errors.Check(errors.WithStack(err))
	count = int(cnt)
	return
}

func (c *CrudRepository[ID, ENTITY]) Exists(ctx context.Context, filter map[string]any) (exists bool, err error) {
	defer errors.Recover(func(e error) { err = e })
	var entity ENTITY
	err = c.query().WithContext(ctx).Where(filter).Select(c.IDField()).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	errors.Check(errors.WithStack(err))
	return true, nil
}

func (c *CrudRepository[ID, ENTITY]) ExistsByID(ctx context.Context, id ID) (exists bool, err error) {
	defer errors.Recover(func(e error) { err = e })
	var entity ENTITY
	err = c.query().WithContext(ctx).Select(c.IDField()).First(&entity, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	errors.Check(errors.WithStack(err))
	return true, nil
}

func (c *CrudRepository[ID, ENTITY]) ExistsByIDs(ctx context.Context, ids []ID) (exists contract.Dict[ID, bool], err error) {
	defer errors.Recover(func(e error) { err = e })
	var entities []ENTITY
	if len(ids) == 0 {
		exists = repository.NewDict[ID, bool](nil)
		return
	}
	err = c.query().WithContext(ctx).Select(c.IDField()).Find(&entities, ids).Error
	errors.Check(errors.WithStack(err))
	exists = repository.NewDictWithSize[ID, bool](len(entities))
	uslice.ForEach(entities, func(item ENTITY) {
		exists.Set(item.GetID(), true)
	})
	return
}

func (c *CrudRepository[ID, ENTITY]) Update(ctx context.Context, filter map[string]any, data map[string]any) (err error) {
	defer errors.Recover(func(e error) { err = e })
	err = c.query().WithContext(ctx).Where(filter).Updates(data).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) UpdateByID(ctx context.Context, id ID, data map[string]any) (err error) {
	defer errors.Recover(func(e error) { err = e })
	err = c.query().WithContext(ctx).Where(map[string]any{c.IDField(): id}).Updates(data).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) UpdateNonZero(ctx context.Context, filter map[string]any, entity ENTITY) (err error) {
	defer errors.Recover(func(e error) { err = e })
	err = c.query().WithContext(ctx).Where(filter).Updates(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) UpdateNonZeroByID(ctx context.Context, id ID, entity ENTITY) (err error) {
	defer errors.Recover(func(e error) { err = e })
	err = c.query().WithContext(ctx).Where(map[string]any{c.IDField(): id}).Updates(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) Delete(ctx context.Context, filter map[string]any) (err error) {
	defer errors.Recover(func(e error) { err = e })
	var entity ENTITY
	err = c.query().WithContext(ctx).Where(filter).Delete(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) DeleteByID(ctx context.Context, id ID) (err error) {
	defer errors.Recover(func(e error) { err = e })
	var entity ENTITY
	err = c.query().WithContext(ctx).Where(map[string]any{c.IDField(): id}).Delete(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) DeleteByIDs(ctx context.Context, ids []ID) (err error) {
	defer errors.Recover(func(e error) { err = e })
	if len(ids) == 0 {
		return
	}
	var entity ENTITY
	err = c.query().WithContext(ctx).Where(c.IDField()+" IN (?)", ids).Delete(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) DeleteAll(ctx context.Context) (err error) {
	defer errors.Recover(func(e error) { err = e })
	var entity ENTITY
	err = c.query().WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) DeleteAllByFilter(ctx context.Context, filter map[string]any) (err error) {
	defer errors.Recover(func(e error) { err = e })
	var entity ENTITY
	err = c.query().WithContext(ctx).Where(filter).Delete(entity).Error
	errors.Check(errors.WithStack(err))
	return
}

func (c *CrudRepository[ID, ENTITY]) DeleteAllByIDs(ctx context.Context, ids []ID) (err error) {
	defer errors.Recover(func(e error) { err = e })
	if len(ids) == 0 {
		return
	}
	var entity ENTITY
	err = c.query().WithContext(ctx).Where(c.IDField()+" IN (?)", ids).Delete(entity).Error
	errors.Check(errors.WithStack(err))
	return
}
