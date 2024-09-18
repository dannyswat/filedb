# FileDB
![Build and Test](https://github.com/dannyswat/filedb/actions/workflows/go.yml/badge.svg)

A filesystem based database with indexing written in Go.

## Feature Overview
- Store objects in file system
- Automatically generate object ID
- Support simple field indexing
- It supports generics

## Limitation
- Not scalable
- Not production tested
- Not support query with value comparation

## Latest Version
v0.2.4-alpha
Please update to the latest version which fixed some bugs.

```console
go get github.com/dannyswat/filedb@v0.2.4-alpha
```

## Future Plan
- Test and fix bugs
- Use it in another project

## Example
```go
package users

import "github.com/dannyswat/filedb"

type UserDB interface {
	Init() error
	GetUserByID(id int) (*User, error)
	GetUserByUserName(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(id int) error
}

type userDB struct {
	db filedb.FileDB[*User]
}

func NewUserDB(path string) UserDB {
	return &userDB{
		db: filedb.NewFileDB[*User](path, []filedb.FileIndexConfig{
			{Field: "UserName", Unique: true},
			{Field: "Email", Unique: true},
		}),
	}
}

func (u *userDB) Init() error {
	return u.db.Init()
}

func (u *userDB) GetUserByID(id int) (*User, error) {
	user, err := u.db.Find(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userDB) GetUserByUserName(username string) (*User, error) {
	users, err := u.db.List("UserName", username)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

func (u *userDB) GetUserByEmail(email string) (*User, error) {
	users, err := u.db.List("Email", email)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

func (u *userDB) CreateUser(user *User) error {
	return u.db.Insert(user)
}

func (u *userDB) UpdateUser(user *User) error {
	return u.db.Update(user)
}

func (u *userDB) DeleteUser(id int) error {
	return u.db.Delete(id)
}

```
