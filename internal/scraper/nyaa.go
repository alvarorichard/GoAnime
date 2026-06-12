package scraper

import (
	"errors"
	"net/http"
)

const (
	nyaaBaseQuerySearch = "https://nyaa.si/?f=0&c=0_0&q="
)

func search(name string) error {
	res, err := http.Get(nyaaBaseQuerySearch + name)
	if err != nil {
		return errors.New("Error doing get request")
	}

}
