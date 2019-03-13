package logic

import (
	"github.com/cohenjo/waste/go/mutators"
)

type ChangeManager struct {
	gorm.Model
	mutators.Change
	Version string
}

func setupChangeManager() {
	db, err := gorm.Open("sqlite3", "waste.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&ChangeManager{})

	// Create
	db.Create(&ChangeManager{Code: "L1212", Price: 1000})

	// Read
	var change ChangeManager
	db.First(&change, 1)                   // find product with id 1
	db.First(&change, "code = ?", "L1212") // find product with code l1212

	// Update - update product's price to 2000
	db.Model(&change).Update("Price", 2000)

	// Delete - delete product
	db.Delete(&change)
}

func mangeChange(change mutators.Change) error {

	// @todo: do we accept change sets? or 1 by 1?

	// @todo: validate change

	// @todo: audit change

	// @todo: schedule change <== should we here or externally?

	return nil
}
