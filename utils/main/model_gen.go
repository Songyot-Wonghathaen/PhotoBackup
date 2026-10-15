package main

import (
	"PhotoVault/repository"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "model/query",                                                      // โฟลเดอร์ปลายทางสำหรับ Query Helper
		ModelPkgPath: "model",                                                            // โฟลเดอร์ปลายทางสำหรับ Struct (Model)
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // โหมดทำงานแบบเรียกใช้งานง่าย
	})

	// 2. ดึงการเชื่อมต่อกับ Database
	db, err := repository.NewDbConnection()
	if err != nil {
		panic(err)
	}

	g.UseDB(db)
	tables, err := db.Migrator().GetTables()
	if err != nil {
		panic(err)
	}
	println("Tables:", tables)

	tags := g.GenerateModel("tags")
	photobase := g.GenerateModel("photos")

	photossRel := gen.FieldRelate(
		field.BelongsTo,
		"PhotoTags",
		photobase,
		&field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"PhotoID"},
				"references": []string{"ID"},
			},
		},
	)

	photoTagsTagRel := gen.FieldRelate(
		field.BelongsTo,
		"Tag",
		tags,
		&field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"TagID"},
				"references": []string{"ID"},
			},
		},
	)

	photoTags := g.GenerateModel("photo_tags", photoTagsTagRel, photossRel)
	photosPhotoTagsRel := gen.FieldRelate(
		field.HasMany,
		"PhotoTags",
		photoTags,
		&field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"PhotoID"},
				"references": []string{"ID"},
			},
		},
	)

	photos := g.GenerateModel("photos", photosPhotoTagsRel)

	g.ApplyBasic(photos, tags, photoTags)

	// Execute การสร้าง code
	g.Execute()

}
