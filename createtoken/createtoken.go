package main

import (
  	"os"
    	"fmt"
    	"log"

    	"golang.org/x/crypto/bcrypt"
)


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide password")
		os.Exit(1)
	}

	password := os.Args[1]

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
	    log.Fatal("Could not generate password:",err)
	}
	
	fmt.Println("Hash:", string(hash))
}
