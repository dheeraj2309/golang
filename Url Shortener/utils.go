package main

import (
	"crypto/rand"
	"math/big"
	"net/url"
	"unicode"
)

// to check if the given url code is valid or not
func isalphanum(s string) bool{
	if len(s) == 0{
		return false
	}
	for _,c := range s{
		if !unicode.IsLetter(c) && !unicode.IsNumber(c){
			return false
		}
	}
	return true
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomCode(length int) (string,error){
	// in go strings are read only and immutable 
	// so we create a new byte buffer to store the code string
	bytes := make([]byte,length)
	// the random function of go expects a bigInt data type to chose the upper boundary
	// to satisfy that need we make the bigInt datatype
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range bytes{
		randIdx,err := rand.Int(rand.Reader, charsetLen)
		if err != nil{
			return "",err
		}
		bytes[i] = charset[randIdx.Int64()]
	}
	return string(bytes),nil
}

func isValidURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	// Ensure absolute URL with valid web scheme and host
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}