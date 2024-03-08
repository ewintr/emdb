package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/go-kit/slugify"
)

const (
	pageTemplate = `+++
title = "{{ .Title }}"
date = {{ .Date }}
draft = false
extra.movie.year = {{ .Year }}
extra.movie.directors = "{{ .Directors }}"
extra.movie.en_title = "{{ .EnTitle }}"
extra.movie.rating = {{ .Rating }}
+++

{{ .Comment }}<!-- more -->`
)

func main() {
	emdb := client.NewEMDB(os.Getenv("EMDB_BASE_URL"), os.Getenv("EMDB_API_KEY"))
	movies, err := emdb.GetMovies()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tpl, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	path := "public"
	Empty(path)

	for _, m := range movies {
		filename := fmt.Sprintf("%s.md", slugify.Slugify(m.EnglishTitle))

		watchedOn := strings.Split(m.WatchedOn, "-")
		watchedOnYear := watchedOn[0]

		if err := os.MkdirAll(fmt.Sprintf("%s/%s", path, watchedOnYear), os.ModePerm); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filePath := fmt.Sprintf("%s/%s/%s", path, watchedOnYear, filename)
		f, err := os.Create(filePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		enTitle := m.EnglishTitle
		if enTitle == m.Title {
			enTitle = ""
		}

		data := struct {
			Title     string
			Date      string
			Year      int
			Directors string
			EnTitle   string
			Rating    string
			Comment   string
		}{
			Title:     m.Title,
			Date:      m.WatchedOn,
			Year:      m.Year,
			Directors: strings.Join(m.Directors, ", "),
			EnTitle:   enTitle,
			Rating:    fmt.Sprintf("%d", m.Rating),
			Comment:   m.Comment,
		}
		if err := tpl.Execute(f, data); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := f.Close(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func Empty(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.RemoveAll(dir + "/" + file.Name())
		if err != nil {
			return err
		}
	}

	return nil
}
