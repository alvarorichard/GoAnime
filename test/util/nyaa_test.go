package test_util

import (
	"io"
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseQueryUrl = "https://nyaa.si/?f=0&c=0_0&q="
)

func TestSearchAnime(t *testing.T) {
	res, err := http.Get(baseQueryUrl + "sono+bisque+doll") 
	if err != nil {
		t.Fatal("Error searching for the anime")
	} else {
		t.Log("Success on searching")
	}

	doc, errDoc := goquery.NewDocumentFromReader(res.Body)
	if errDoc != nil {
		t.Fatal("Error creating goquery document from request body")
	}

	anchors := doc.Find("a[title][href]")

	anchors.Each(func(i int, s *goquery.Selection) {
			// iterate over all attributes of the anchor node
			for _, attr := range anchors.Nodes[i].Attr {
				t.Logf("attr: %s=%s", attr.Key, attr.Val)
			}
	})
}

func TestGetHome(t *testing.T) {
	res, err := http.Get("https://nyaa.si")
	if err != nil {
		t.Fatal("Error getting the nyaa home")
	} else {
		t.Log("Success on geting nyaa home page")
	}

	bodyBytes, errGetBody := io.ReadAll(res.Body)
	if errGetBody != nil {
		t.Fatal("Error reading the request answer body")
	}
	_ = bodyBytes

}

