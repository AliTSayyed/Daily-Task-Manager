package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// create our custom data type called Todo
type Todo struct {
	ID        int    `json:"_id" bson:"_id"` // mongo db stores data as bson (binary json)
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
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
	app := fiber.New()

	// Define routes for CRUD operations on the "todos" resource
	// GET /api/todos - Retrieve all todos
	app.Get("/api/todos", getTodos)

	// POST /api/todos - Create a new todo
	//app.Post("/api/todos", createTodo)

	// PATCH /api/todos/:id - Update a specific todo by its ID
	//app.Patch("/api/todos/:id", updateTodo)

	// DELETE /api/todos/:id - Delete a specific todo by its ID
	//app.Delete("/api/todos/:id", deleteTodo)

	// Get the port from environment variables, default to "5000" if not set
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}
	// Start the Fiber application and listen on the specified port
	log.Fatal(app.Listen("0.0.0.0:" + PORT))
}

func getTodos(c *fiber.Ctx) error {
	// create list to store all the todos
	var todos []Todo
	// create a cursor object that points the todos in the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal("Couldn't get todo list:", err)
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
	return c.JSON(todos)
}

// func createTodos(c *fiber.Ctx) error{}
// func updateTodos(c *fiber.Ctx) error{}
// func deleteTodos(c *fiber.Ctx) error{}
