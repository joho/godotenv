package main

import (
	"fmt"
	"github.com/open-source/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load("trytry")
	if err != nil {
		log.Fatal("FATAL ERROR: ", err)
	}

	//err = godotenv.Load("deploy.env", "dev.env")
	//if err != nil {
	//	log.Fatal("FATAL ERROR: ", err)
	//}

	dbA := os.Getenv("A")
	dbB := os.Getenv("B")
	dbC := os.Getenv("C")
	dbD := os.Getenv("D")
	dbName := os.Getenv("NAME")
	fmt.Println(dbA, "...", dbB, "...", dbC, "...", dbD)
	fmt.Println(dbName)

	dbAA := os.Getenv("AA")
	dbBB:=  os.Getenv("BB")
	dbCC := os.Getenv("CC")
	dbDD := os.Getenv("DD")
	fmt.Println(dbAA, "...", dbBB, "...", dbCC, "...", dbDD)

	dbAAA := os.Getenv("AAA")
	dbBBB := os.Getenv("BBB")
	dbCCC := os.Getenv("CCC")
	dbDDD := os.Getenv("DDD")
	fmt.Println(dbAAA, "...", dbBBB, "...", dbCCC, "...", dbDDD)



	// now do something with s3 or whatever
}
