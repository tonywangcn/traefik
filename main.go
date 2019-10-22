package main

import (
	"fmt"
	"net/http"
	"golang.org/x/text/language"
)

func main() {
	http.HandleFunc("/", index) 
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
		fmt.Println(err)
    }
}

handler(w http.ResponseWriter, r *http.Request) {
    lang, _ := r.Cookie("lang")
    accept := r.Header.Get("Accept-Language")
    tag, _ := language.MatchStrings(matcher, lang.String(), accept)

    // tag should now be used for the initialization of any
    // locale-specific service.
}