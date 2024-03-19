package main

import (
	mw "GORM/BasicMiddleware"
	ent "GORM/Event"
	logging "GORM/MiddleWare"
	cache "GORM/REDIS"
	ses "GORM/SessionManage"

	jwt "GORM/Token"

	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

var Client *redis.Client

func main() {

	ent.InitDB()

	Client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	//Setting the client
	cache.SetRedisClient(Client)

	r := mux.NewRouter()

	//for logging MiddleWare
	r.Use(logging.Logging)

	//for redisCaching MiddleWare requests for only get--method.
	r.Use(cache.RedisCaching)

	//Creating RESTAPI's
	r.Handle("/addProduct", mw.NewAuth(http.HandlerFunc(ent.AddProduct))).Methods("POST")
	r.Handle("/getAllProducts", mw.NewAuth(http.HandlerFunc(ent.GetAllProducts))).Methods("GET")
	r.Handle("/getProduct/{id}", mw.NewAuth(http.HandlerFunc(ent.GetProduct))).Methods("GET")
	r.Handle("/updateProduct/{id}", mw.NewAuth(http.HandlerFunc(ent.UpdateProduct))).Methods("PUT")
	r.Handle("/deleteProduct/{id}", mw.NewAuth(http.HandlerFunc(ent.DeleteProduct))).Methods("DELETE")

	//JWT-TOKEN
	r.HandleFunc("/jwt-token", jwt.Token)

	// rdb := redis.InitRedis()

	//for Session Management
	r.HandleFunc("/session-login", ses.LoginSession)
	r.HandleFunc("/protected", ses.Protected)
	r.HandleFunc("/logout", ses.Logout)

	//Activating the Server
	// http.ListenAndServe(":8081", r)
	http.ListenAndServe("0.0.0.0:8082", r)
}
