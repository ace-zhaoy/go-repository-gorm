package repository_gorm

import (
	"context"
	"github.com/ace-zhaoy/errors"
	"github.com/ace-zhaoy/go-repository"
	"github.com/ace-zhaoy/go-repository/contract"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"log"
)

// table: users
type UserRepository struct {
	contract.CrudRepository[int64, *User]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		CrudRepository: NewCrudRepository[int64, *User](db),
		db:             db,
	}
}

func ExampleCrudRepository() {
	defer errors.Recover(func(e error) { log.Fatalf("%+v", e) })
	var ctx = context.Background()
	var db *gorm.DB
	userRepository := NewUserRepository(db)
	id, err := userRepository.Create(ctx, &User{
		ID:   idGen.Generate(),
		Name: "test",
	})
	errors.Check(err)
	user, err := userRepository.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		// TODO: handle not found
		return
	}
	errors.Check(err)
	_ = user
}

type Role struct {
	ID        int64                 `json:"id"`
	Name      string                `json:"name"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at"`
}

func (u *Role) GetID() int64 {
	return u.ID
}

func (u *Role) SetID(id int64) {
	u.ID = id
}

// table: role
type RoleRepository struct {
	contract.CrudRepository[int64, *Role]
	db *gorm.DB
}

func (r *RoleRepository) Table() string {
	return "role"
}

func (r *RoleRepository) OtherMethod() {

}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	r := &RoleRepository{}
	r.db = db.Table(r.Table())
	r.CrudRepository = NewCrudRepository[int64, *Role](r.db)

	return r
}

func ExampleCrudRepository_SoftDelete() {
	defer errors.Recover(func(e error) { log.Fatalf("%+v", e) })
	var ctx = context.Background()
	var db *gorm.DB
	roleRepository := NewRoleRepository(db)
	id, err := roleRepository.Create(ctx, &Role{
		ID:   idGen.Generate(),
		Name: "test",
	})
	errors.Check(err)
	err = roleRepository.DeleteByID(ctx, id)
	errors.Check(err)

	// Find deleted data
	role, err := roleRepository.Unscoped().FindByID(ctx, id)
	// role exists, and DeletedAt > 0
	_ = role
}
