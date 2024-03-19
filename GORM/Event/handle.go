package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// model
type Product struct {
	gorm.Model           //To Create additional Fields && Model is a --struct
	Item         string  `json:"item"`
	Price        float32 `json:"price"`
	ReturnPolicy int     `json:"return_policy"`
}

var wg sync.WaitGroup
var db *gorm.DB

func InitDB() {
	dsn := "root:Sathyabama*40110529@tcp(host.docker.internal:3306)/product?parseTime=true"
	// dsn := "root:Sathyabama*40110529@tcp(localhost:3306)/product?parseTime=true"

	//error is interface
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Print(err.Error())
		panic("Unable to Connect")
	}

	//Creating the Table in Schema
	errr := db.AutoMigrate(&Product{})
	if errr != nil {
		fmt.Println("Error migrating database:", errr)
		panic("Unable to migrate database")
	}
}
func AddProduct(w http.ResponseWriter, r *http.Request) {
	//Create a Var
	var prod Product

	//requesting to Give response in JSON format
	w.Header().Add("Content-Type", "application/json")

	//JSON-to-Struct
	decode := json.NewDecoder(r.Body) //reads the data which is present in JSON
	err := decode.Decode(&prod)       //Storing in a variable
	if err != nil {
		http.Error(w, "Enter the data in Correct Format", http.StatusNotAcceptable)
		return
	}
	if db == nil {
		http.Error(w, "Database connection is not initialized", http.StatusInternalServerError)
		return
	}

	//Go-rotunies  -- Maintained by run-time
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Save the product
		result := db.Save(&prod)
		if result.Error != nil {
			http.Error(w, "Failed to save product", http.StatusInternalServerError)
			return
		}
		// Respond with the saved product
		json.NewEncoder(w).Encode(prod)

	}()
	wg.Wait()

}

func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	//Slice
	var prod []Product
	wg.Add(1)
	go func() {
		// retrieve records from the database that match certain criteria and Store in &prod --- instance of a struct
		defer wg.Done()
		_ = db.Find(&prod)
	}()
	wg.Wait()
	//requesting to Give response in JSON format
	w.Header().Add("Content-Type", "application/json")
	//Encode
	json.NewEncoder(w).Encode(prod) //Encodes -- Slice data str to JSON Format
}

func GetProduct(w http.ResponseWriter, r *http.Request) {
	var prod Product
	wg.Add(1)
	go func() {
		defer wg.Done()

		vars := mux.Vars(r)
		id := vars["id"]

		db.Where("id=?", id).First(&prod) //Here "i=?"d should be same as in Table because it is PK
	}()
	wg.Wait()
	//requesting to Give response in JSON format
	w.Header().Add("Content-Type", "application/json")

	//If roduct is t Present.
	if prod.ID == 0 { // Assuming ID is of type int, change this condition if needed
		// Product with given ID not found
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}

	json.NewEncoder(w).Encode(prod)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	//variables from the request
	vars := mux.Vars(r)
	id := vars["id"]
	var (
		prodChan   = make(chan bool)
		updateChan = make(chan bool)
		saveChan   = make(chan bool)
		notFound   bool
		notJSON    bool
	)
	var existingProd Product

	wg.Add(1)
	go func() {
		defer func() { prodChan <- true }()
		//to check Whether the product is present are not and Storing in a new varible
		result := db.Where("id=?", id).First(&existingProd)
		if result.RowsAffected == 0 { //DB is a struct which contains ERROR
			notFound = true
			http.Error(w, "Product is not Present", http.StatusNotFound)
			return
		}
	}()

	wg.Add(1)
	//var
	var update Product
	go func() {
		defer func() { updateChan <- true }()
		//Decoding and storing in a new Variable
		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			notJSON = true
			http.Error(w, "Not in JSON FORMAT", http.StatusBadRequest)
			return
		}
	}()

	log.Println(existingProd)

	wg.Add(1)

	go func() {
		//Waiting for this gorotunies to complete
		<-prodChan
		<-updateChan

		if notFound || notJSON {
			saveChan <- true // Signal completion to unblock the waiting routine
			return
		}

		//Updating
		// existingProd.Model = update.Model
		existingProd.Item = update.Item
		existingProd.Price = update.Price
		existingProd.ReturnPolicy = update.ReturnPolicy
		// log.Println(update)

		res := db.Save(&existingProd)
		if res.Error != nil { //DB is a struct which contains ERROR
			http.Error(w, "Unable to update the Product", http.StatusInternalServerError)
			return
		}
		saveChan <- true
		w.Header().Add("Content-Type", "application/json")
		//For Output
		json.NewEncoder(w).Encode(existingProd)
	}()
	//Waiting for the SaveChan
	<-saveChan
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var prod Product

	wg.Add(1)
	go func() {
		defer wg.Done()

		result := db.Where("id=?", id).First(&prod)
		if result.RowsAffected == 0 {
			http.Error(w, "Product is not present", http.StatusNotFound)
			return
		}
		res := db.Delete(&prod)
		if res.Error != nil {
			http.Error(w, "Unable to delete Product", http.StatusInternalServerError)
			return
		}
		//Response
		json.NewEncoder(w).Encode(&prod)
	}()
	wg.Wait()

}
