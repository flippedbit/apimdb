package apimdb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

const imdbUserURL = "https://www.imdb.com/name/"

type IMDBUser struct {
	name     string
	id       string
	knownFor []string
	imdbBody string
}

func NewIMDBUser() *IMDBUser {
	return &IMDBUser{
		name: "",
		id:   "",
	}
}

func (person *IMDBUser) GetPersonIDByName(name string) (string, error) {
	url := imdbSearchURL + "nm&q=" + strings.ReplaceAll(name, " ", "+")
	if person.imdbBody == "" {
		req, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer req.Body.Close()

		if body, err := ioutil.ReadAll(req.Body); err == nil {
			person.imdbBody = string(body)
		}
	}

	reader := strings.NewReader(person.imdbBody)
	z := html.NewTokenizer(reader)

	result := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "td" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "result_text" {
						result = true
					}
				}
			} else if t.Data == "a" && result == true {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						if href, err := splitIMDBName(attr.Val); err == nil {
							person.id = href
							return href, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("Could not find person")
}

func (person *IMDBUser) GetPersonByID(id string) error {
	person.id = id
	url := imdbUserURL + id
	if person.imdbBody == "" {
		req, err := http.Get(url)
		if err != nil {
			return err
		}
		defer req.Body.Close()

		if body, err := ioutil.ReadAll(req.Body); err == nil {
			person.imdbBody = string(body)
		} else {
			return err
		}
	}

	if person.name == "" {
		if err := person.fetchName(); err != nil {
			return err
		}
	}

	if len(person.knownFor) == 0 {
		if err := person.fetchKnownFor(); err != nil {
			return err
		}
	}

	return nil
}

func (person *IMDBUser) fetchName() error {
	reader := strings.NewReader(person.imdbBody)
	z := html.NewTokenizer(reader)

	header := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break
		case html.TextToken:
			if header {
				person.name = string(z.Text())
				return nil
			} else {
				continue
			}
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "h1" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "header" {
						header = true
					}
				}
			}
		}
	}
	return nil
}

func (person IMDBUser) Name() string {
	return person.name
}

func (person *IMDBUser) fetchKnownFor() error {
	reader := strings.NewReader(person.imdbBody)
	z := html.NewTokenizer(reader)

	knownDiv := false
	depth := 0
	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "a" && len(t.Attr) > 0 {
				href := ""
				known := false
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "knownfor-ellipsis" {
						known = true
					} else if attr.Key == "href" {
						href = attr.Val
					}
				}
				if known {
					if name, err := splitIMDBName(href); err == nil {
						person.knownFor = append(person.knownFor, name)
						known = false
					}
				}
			} else if t.Data == "div" && len(t.Attr) > 0 {
				if knownDiv {
					depth++
				} else {
					for _, attr := range t.Attr {
						if attr.Key == "id" && attr.Val == "knownfor" {
							depth++
							knownDiv = true
						}
					}
				}
			}
		} else if tt == html.EndTagToken {
			if knownDiv == true {
				if tn, _ := z.TagName(); string(tn) == "div" {
					depth--
					if depth == 0 {
						break
					}
				}
			}
		}
	}
	if len(person.knownFor) == 0 {
		return fmt.Errorf("Could not find known for")
	} else {
		return nil
	}
}

func (person IMDBUser) KnownFor() []string {
	return person.knownFor
}

func (person IMDBUser) ID() string {
	return person.id
}
