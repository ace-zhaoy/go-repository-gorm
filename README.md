# 介绍
这是一个使用`gorm`作为驱动的 repository 库，支持 [github.com/ace-zhaoy/go-repository](https://github.com/ace-zhaoy/go-repository) 协议。

# 使用
```shell
go get github.com/ace-zhaoy/go-repository-gorm
```
## 快速使用
> 假设表名为`users`
```go
package main

import (
	"context"
	"fmt"
	goid "github.com/ace-zhaoy/go-id"
	repositorygorm "github.com/ace-zhaoy/go-repository-gorm"
	"github.com/ace-zhaoy/go-repository/contract"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (u *User) GetID() int64 {
	return u.ID
}

func (u *User) SetID(id int64) {
	u.ID = id
}

type UserRepository struct {
	contract.CrudRepository[int64, *User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		CrudRepository: repositorygorm.NewCrudRepository[int64, *User](db),
	}
}

func main() {
	rootPassword := ""
	mysqlEndpoint := ""
	databaseName := ""
	dsn := fmt.Sprintf("root:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", rootPassword, mysqlEndpoint, databaseName)
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	userRepository := NewUserRepository(db)
	id, err := userRepository.Create(context.Background(), &User{
		ID:   goid.GenID(),
		Name: "test",
	})
	fmt.Printf("id = %d, err = %+v\n", id, err)
}

```

> 假设表名为`user`
```go
package main

import (
	"context"
	"fmt"
	goid "github.com/ace-zhaoy/go-id"
	repositorygorm "github.com/ace-zhaoy/go-repository-gorm"
	"github.com/ace-zhaoy/go-repository/contract"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (u *User) GetID() int64 {
	return u.ID
}

func (u *User) SetID(id int64) {
	u.ID = id
}

type UserRepository struct {
	contract.CrudRepository[int64, *User]
}

func (u *UserRepository) Table() string {
	return "user"
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	u := &UserRepository{}
	u.CrudRepository = repositorygorm.NewCrudRepository[int64, *User](db.Table(u.Table()))
	return u
}

func main() {
	rootPassword := ""
	mysqlEndpoint := ""
	databaseName := ""
	dsn := fmt.Sprintf("root:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", rootPassword, mysqlEndpoint, databaseName)
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	userRepository := NewUserRepository(db)
	id, err := userRepository.Create(context.Background(), &User{
		ID:   goid.GenID(),
		Name: "test",
	})
	fmt.Printf("id = %d, err = %+v\n", id, err)
}
```

## 使用软删
> 增加 DeletedAt 属性即可
```go
package main

import (
	"context"
	"fmt"
	"github.com/ace-zhaoy/errors"
	goid "github.com/ace-zhaoy/go-id"
	"github.com/ace-zhaoy/go-repository"
	repositorygorm "github.com/ace-zhaoy/go-repository-gorm"
	"github.com/ace-zhaoy/go-repository/contract"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"log"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at"`
}

func (u *User) GetID() int64 {
	return u.ID
}

func (u *User) SetID(id int64) {
	u.ID = id
}

type UserRepository struct {
	contract.CrudRepository[int64, *User]
}

func (u *UserRepository) Table() string {
	return "user"
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	u := &UserRepository{}
	u.CrudRepository = repositorygorm.NewCrudRepository[int64, *User](db.Table(u.Table()))
	return u
}

func main() {
	defer errors.Recover(func(e error) { 
		log.Fatalf("err: %+v\n", e)
	})
	rootPassword := ""
	mysqlEndpoint := ""
	databaseName := ""
	dsn := fmt.Sprintf("root:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", rootPassword, mysqlEndpoint, databaseName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	errors.Check(err)
	userRepository := NewUserRepository(db)
	ctx := context.Background()
	id, err := userRepository.Create(ctx, &User{
		ID:   goid.GenID(),
		Name: "test",
	})
	errors.Check(err)

	err = userRepository.DeleteByID(ctx, id)
	errors.Check(err)
	
	// 常规查找（无法找到被删除的数据）
	_, err = userRepository.FindByID(ctx, id)
	fmt.Println(errors.Is(err, repository.ErrNotFound))
	
	// 使用 Unscoped 查找（可以找到被删除的数据）
	user, err := userRepository.Unscoped().FindByID(ctx, id)
	errors.Check(err)
	fmt.Println(user.DeletedAt)
}

```

## 自定义方法
> 例如增加 OtherMethod 方法
```go
package main

import (
	"fmt"
	"github.com/ace-zhaoy/errors"
	repositorygorm "github.com/ace-zhaoy/go-repository-gorm"
	"github.com/ace-zhaoy/go-repository/contract"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"log"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at"`
}

func (u *User) GetID() int64 {
	return u.ID
}

func (u *User) SetID(id int64) {
	u.ID = id
}

type UserRepository struct {
	contract.CrudRepository[int64, *User]
	db *gorm.DB
}

func (u *UserRepository) Table() string {
	return "user"
}

func (u *UserRepository) connect() *gorm.DB {
	if u.IsUnscoped() {
		return u.db.Unscoped()
	}
	return u.db
}

func (u *UserRepository) OtherMethod() {
	// r.connect().xxx
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	u := &UserRepository{}
	u.db = db.Table(u.Table())
	u.CrudRepository = repositorygorm.NewCrudRepository[int64, *User](u.db)
	return u
}

func main() {
	defer errors.Recover(func(e error) { 
		log.Fatalf("err: %+v\n", e)
	})
	rootPassword := ""
	mysqlEndpoint := ""
	databaseName := ""
	dsn := fmt.Sprintf("root:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", rootPassword, mysqlEndpoint, databaseName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	errors.Check(err)
	userRepository := NewUserRepository(db)
	userRepository.OtherMethod()
}

```