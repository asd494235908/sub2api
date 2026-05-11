package main

import (
  "fmt"
  "golang.org/x/crypto/bcrypt"
)

func main() {
  hashed, err := bcrypt.GenerateFromPassword([]byte("admin123456"), bcrypt.DefaultCost)
  if err != nil {
    panic(err)
  }
  fmt.Print(string(hashed))
}
