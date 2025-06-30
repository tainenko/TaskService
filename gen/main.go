package main

import (
	"github/TaskService/model"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./dao",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode,
		WithUnitTest:      true,
		FieldNullable:     true, // generate pointer when field is nullable
		FieldCoverable:    true, // generate pointer when field has default value, to fix problem zero value cannot be assign: https://gorm.io/docs/create.html#Default-Values
		FieldSignable:     true, // detect integer field's unsigned type, adjust generated data type
		FieldWithIndexTag: true, // generate with gorm index tag
		FieldWithTypeTag:  true, // generate with gorm column type tag
	})

	gormdb, _ := gorm.Open(postgres.Open("postgres://postgres:password@localhost:5432/postgres"))
	g.UseDB(gormdb) // reuse your gorm db

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	g.ApplyBasic(model.Task{})

	// Generate the code
	g.Execute()
}
