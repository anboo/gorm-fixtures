## Gorm fixtures loader
Only for MySQL or PostgreSQL

## Why?
Our project is a fixture loader for GORM, a popular ORM library for Go. It allows developers to define fixtures, which are sets of pre-defined data, and load them into the database for testing purposes. Fixtures can represent various entities in the database, such as users, products, or orders.

The fixture loader provides functionalities to load specific fixtures or their dependencies, making it suitable for integration testing scenarios. For instance, before running integration tests, developers can use the fixture loader to populate the database with test data by loading all fixtures or only specific ones relevant to the test case. After the test execution, the fixture loader can also be used to clean up the database by removing the test data.

By integrating this fixture loader into our testing workflow, we ensure that our integration tests run on a consistent and predictable database state, helping us to identify potential issues and ensure the reliability of our applications. Additionally, it facilitates rapid development and testing iterations by automating the setup and teardown of test data.

## Install
`go get github.com/anboo/gorm-fixtures@latest`

## Usage:
```go
package main

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/anboo/gorm-fixtures"
)

// User represents the user model.
type User struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"not null"`
	Age  uint   `gorm:"not null"`
}

// AccessToken represents the access token model.
type AccessToken struct {
	ID        uint   `gorm:"primarykey"`
	Token     string `gorm:"not null"`
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// UserFixture represents the fixture for loading users into the database.
type UserFixture struct{}

// Load loads users into the database.
func (f *UserFixture) Load(ctx *gorm_fixtures.LoadCtx, db *gorm.DB) error {
	users := []User{
		{ID: 1, Name: "Alice", Age: 30},
		{ID: 2, Name: "Bob", Age: 35},
		{ID: 3, Name: "Charlie", Age: 25},
	}

	// Save users to the database
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return err
		}

		// Save a reference to the created user in the context
		ctx.SetReference(fmt.Sprintf("user:%d", user.ID), user)
	}

	return nil
}

// GetRequiredRelations returns an empty list of dependencies,
// as UserFixture does not depend on other fixtures.
func (f *UserFixture) GetRequiredRelations() []gorm_fixtures.Fixture {
	return []gorm_fixtures.Fixture{}
}

// Name returns the fixture name.
func (f *UserFixture) Name() string {
	return "UserFixture"
}

// AccessTokenFixture represents the fixture for loading access tokens into the database.
type AccessTokenFixture struct{}

// Load loads access tokens into the database.
func (f *AccessTokenFixture) Load(ctx *gorm_fixtures.LoadCtx, db *gorm.DB) error {
	// Get a reference to the user from the context
	user, err := ctx.GetReference("user:1")
	if err != nil {
		return err
	}

	// Create an access token for the user
	accessToken := AccessToken{
		Token:  "abc123",
		UserID: user.(User).ID,
	}

	// Save the access token to the database
	if err := db.Create(&accessToken).Error; err != nil {
		return err
	}

	return nil
}

// GetRequiredRelations returns a list of dependencies,
// as AccessTokenFixture depends on UserFixture.
func (f *AccessTokenFixture) GetRequiredRelations() []gorm_fixtures.Fixture {
	return []gorm_fixtures.Fixture{&UserFixture{}}
}

// Name returns the fixture name.
func (f *AccessTokenFixture) Name() string {
	return "AccessTokenFixture"
}

func main() {
	// Connect to the PostgreSQL database
	dsn := "host=localhost user=your_user password=your_password dbname=your_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Apply migrations to create tables
	db.AutoMigrate(&User{}, &AccessToken{})

	// Create an instance of FixtureLoader
	fixtureLoader := gorm_fixtures.NewFixtureLoader(db, &UserFixture{}, &AccessTokenFixture{})

	// Configuration for loading fixtures
	config := gorm_fixtures.Config{
		ShowProgressBar:     true,
		ResetAutoIncrements: false,
		TruncateAllTables:   true,
	}

	// Load fixtures
	if err := fixtureLoader.Load(context.Background(), config); err != nil {
		panic(err)
	}

	fmt.Println("Fixtures loaded successfully!")
}
```

### Load one fixture with all dependencies
```go
fixtureLoader := gorm_fixtures.NewFixtureLoader(db, &UserFixture{}, &AccessTokenFixture{})
fixtureLoader.LoadFixtureByName("AccessTokenFixture")
```

### Or load specific struct fixture
```go
fixtureLoader.LoadFixture(&AccessTokenFixture{})
```