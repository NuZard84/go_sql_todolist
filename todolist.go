package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

func getLists(w http.ResponseWriter, r *http.Request) {

	log.Info("API getlist OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func getItemById(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoItems not found in DB")
		return false
	}
	return true
}

func getTodoItems(completed bool) interface{} {

	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	log.WithFields(log.Fields{"description": description}).Info("Add new TodoItem. Saving to database.")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Value)
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	Id, _ := strconv.Atoi(vars["Id"])
	isItemExist := getItemById(Id)

	if !isItemExist {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"message": "Item not found"}`)
		return
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": Id, "completed": completed}).Info("Update TodoItem. Saving to database.")
		todo := &TodoItemModel{}
		db.First(&todo, Id)
		todo.Completed = completed
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"message": "Item updated successfully"}`)
	}

}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["Id"])

	isItemExist := getItemById(id)
	if !isItemExist {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"message": "Item not found"}`)
		return
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Delete TodoItem. Saving to database.")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"message": "Item deleted successfully"}`)
	}
}

func GetCompletedITems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get completed TodoItems")
	completedTodoItems := getTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func GetIncompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get incompleted TodoItems")
	incompletedTodoItems := getTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incompletedTodoItems)
}

// It runs automatically once per package, even before the main() function is executed.
// It is used to set up the logger or db connection and more.
func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	var err error
	db, err = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

	if err != nil {
		log.Fatalf("Failed to connect to database %v", err)
	}
	log.Info("Database connection successful")
}

func main() {
	defer db.Close()

	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	fmt.Println("Starting the application...")
	log.Info("Starting the application server...")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", getLists).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo-completed", GetCompletedITems).Methods("GET")
	router.HandleFunc("/todo-incomplete", GetIncompletedItems).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)
	http.ListenAndServe(":8000", handler)
}
