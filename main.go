package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// create our custom data type called Todo
type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // mongo db stores data as bson (binary json)
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("hello world")

	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Get MongoDB URI from environment variables
	MONGODB_URI := os.Getenv("MONGODB_URI")

	// Set up MongoDB client options using the URI
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	// clientOptions now contains configuration settings such as the MongoDB server address

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}

	// disconenct the database after main function is done
	defer client.Disconnect(context.Background())

	// Ping MongoDB to verify the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Could not ping to database:", err)
	}

	fmt.Println("Connected to MONGODB ATLAS")

	// Access the "todos" collection in the "golang_db" database
	collection = client.Database("golang_db").Collection("todos")

	// Initialize a new Fiber application instance
	router := fiber.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))

	// Connecting backend to front end using cors from the fiber framework

	// Define routes for CRUD operations on the "todos" resource
	// GET /api/todos - Retrieve all todos
	router.Get("/api/todos", getTodos)
	// POST /api/todos - Create a new todo
	router.Post("/api/todos", createTodos)
	// PATCH /api/todos/:id - Update a specific todo by its ID
	router.Patch("/api/todos/:id", updateTodos)
	// DELETE /api/todos/:id - Delete a specific todo by its ID
	router.Delete("/api/todos/:id", deleteTodos)

	// Get the port from environment variables, default to "5000" if not set
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}

	// Start the Fiber application and listen on the specified port
	log.Fatal(router.Listen("0.0.0.0:" + PORT))
}

func getTodos(c *fiber.Ctx) error {
	// create list to store all the todos
	var todos []Todo
	// create a cursor object that points the todos in the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Printf("Couldn't get todo list:")
		return err
	}

	// close cursor at the end of the get function
	defer cursor.Close(context.Background())

	// go through each correct item in the database and append it to the todos list
	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}
	// return JSON version of todo (default completed status is false)
	return c.JSON(todos)
}

func createTodos(c *fiber.Ctx) error {
	todo := new(Todo)
	// binds the request body to a struct
	if err := c.BodyParser(todo); err != nil {
		log.Printf("Could not parse body:")
		return err
	}

	// can not pass an empty struct
	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body is required"})
	}

	// if it is a valid todo, insert it into mongo db collection
	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		log.Printf("Could not insert todo:")
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	// need to convert hex id back to int id
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	// update todo using BSON formated changes
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Could not update todo")
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Could not delete todo")
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": true})
}
