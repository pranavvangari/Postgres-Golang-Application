package main

import (
	//general imports
	"fmt"
	"log"

	//import for fiber
	"github.com/gofiber/fiber/v2"

	//imports for gorm and postgres db
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Entity2 struct {
	ID          uint   `gorm:"type:serial;primaryKey"`
	Type        string `gorm:"not null"`
	Name        string `gorm:"not null"`
	Description string `gorm:"not null"`
}

type Entity struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TextResponse struct {
	Message string `json:"message"`
}

var db *gorm.DB

func main() {

	// Code for connecting to the postgres db
	creds := "host=localhost user=postgres password=mysecretpassword dbname=realentities port=5432 sslmode=disable"

	var err error
	db, err = gorm.Open(postgres.Open(creds), &gorm.Config{})

	if err != nil {
		log.Fatalf("failed to conn to database: %v", err)
	}

	fmt.Println("connected to postgresql db")

	err = db.AutoMigrate(&Entity2{})

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database migration successful")

	// Code for routes

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error { // c for context
		return c.JSON("hello world!")
	})

	entitiesAPI := app.Group("/entities")

	entitiesAPI.Post("/", addEntity)
	entitiesAPI.Get("/", getEntities)
	entitiesAPI.Get(":id", getSpecificEntity)
	entitiesAPI.Put("/", updateEntity)
	entitiesAPI.Delete(":id", deleteEntity)

	//log.Fatal(app.Listen(":8080"))
	app.Listen(":8080")
}

func addEntity(c *fiber.Ctx) error {

	var entity Entity
	var entity2 Entity2

	if err := c.BodyParser(&entity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	entity2.Name = entity.Name
	entity2.Type = entity.Type
	entity2.Description = entity.Description

	result := db.Create(&entity2) //inserts a new entity into the database

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.JSON(entity)

}

func getEntities(c *fiber.Ctx) error {
	var entities []Entity2

	result := db.Find(&entities)

	fmt.Println(entities)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.JSON(entities)
}

func getSpecificEntity(c *fiber.Ctx) error {
	entityID := c.Params("id")

	var entity Entity2

	result := db.Find(&entity, entityID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Entity not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if entity.Description == "" {
		errMsg := TextResponse{Message: "entity does not exist"}
		return c.Status(fiber.StatusNotFound).JSON(errMsg)
	}

	return c.JSON(entity)
}

func updateEntity(c *fiber.Ctx) error {
	var newEntity Entity2

	err := c.BodyParser(&newEntity)

	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(TextResponse{Message: "status conflict"})
	}

	var existingEntity Entity2

	result := db.First(&existingEntity, newEntity.ID)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(TextResponse{Message: "result did not work"})
	}

	existingEntity.Type = newEntity.Type
	existingEntity.Description = newEntity.Description
	existingEntity.Name = newEntity.Name

	result = db.Save(&existingEntity)

	if result.Error != nil {
		return c.Status(fiber.StatusConflict).JSON(TextResponse{Message: "could not save to db"})
	}

	return c.JSON(existingEntity)

}

func deleteEntity(c *fiber.Ctx) error {
	entityID := c.Params("id")

	var entity Entity2

	result := db.Find(&entity, entityID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(TextResponse{Message: ".find() did not work"})
	}

	result = db.Delete(&entity)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(TextResponse{Message: "delete did not work"})
	}

	return c.JSON(entity)

}
