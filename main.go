package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"regexp"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var regex *regexp.Regexp

func main() {

	regex = regexp.MustCompile(`(^[0-9]+)\.png$`)

	dsn := "root:root@tcp(127.0.0.1:3306)/corona?charset=utf8mb4&parseTime=True&loc=Local"
	var err error

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	migrate()

	router := httprouter.New()
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			header := w.Header()
			header.Set("Access-Control-Allow-Methods", r.Header.Get("Allow"))
			header.Set("Access-Control-Allow-Origin", "http://localhost:4200")
			header.Set("Access-Control-Allow-Headers", "Authorization, Accept, Content-Type, Content-Length, Accept-Encoding")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	jwks := "https://coronamans.auth0.com/.well-known/jwks.json"
	audience := "https://coronamans.pharmatics.io:8443"
	issuer := "https://coronamans.eu.auth0.com/"

	corsSpec := CORSOpts{
		AllowedOriginPatterns: []string{"http://localhost:4200"},
		AllowedMethods:        []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS"},
	}

	auth := WithAuthSigningMethodRS256(jwks, audience, issuer)
	cors := WithCORS(corsSpec)

	chain := NewChain(cors, auth)

	router.GET("/barcode/:image", GetBarcodeImage)

	router.POST("/employee", chain.Then(CreateEmployee))
	router.GET("/employee/:barcode", chain.Then(GetEmployee))
	router.GET("/employees", chain.Then(GetEmployees))
	router.DELETE("/employee/:barcode", chain.Then(DeleteEmployee))

	router.POST("/check/:barcode", chain.Then(Check))
	router.GET("/status/:barcode", chain.Then(GetStatus))

	router.GET("/generate.png", chain.Then(Generate))

	log.Printf("start web server :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func migrate() {
	db.AutoMigrate(&Attendance{}, &Employee{})
	db.Exec("ALTER DATABASE corona CHARACTER SET utf8 COLLATE utf8_general_ci;")
	db.Exec("ALTER TABLE employees CHARACTER SET utf8 COLLATE utf8_general_ci;")
	db.Exec("ALTER TABLE name DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci")
	startID := time.Now().Unix()
	query := fmt.Sprintf("CREATE SEQUENCE employee_serial INCREMENT 13 START %d", startID)
	db.Exec(query)
}