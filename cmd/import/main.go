package main

import (
	"fmt"
	"os"

	"code.ewintr.nl/emdb/cmd/api-service/moviestore"
	"code.ewintr.nl/emdb/storage"
)

func main() {
	dbSQLite, err := moviestore.NewSQLite("./emdb.db")
	if err != nil {
		fmt.Printf("could not create new sqlite repo: %s", err.Error())
		os.Exit(1)
	}

	pgConnStr := ""
	dbPostgres, err := storage.NewPostgres(pgConnStr)
	if err != nil {
		fmt.Printf("could not create new postgres repo: %s", err.Error())
		os.Exit(1)
	}

	//fmt.Println("movies")
	//movieRepoSqlite := moviestore.NewMovieRepository(dbSQLite)
	//movieRepoPG := moviestore.NewMovieRepositoryPG(dbPostgres)
	//
	//movies, err := movieRepoSqlite.FindAll()
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//for _, movie := range movies {
	//	if err := movieRepoPG.Store(movie); err != nil {
	//		fmt.Println(err)
	//		os.Exit(1)
	//	}
	//}
	fmt.Println("reviews")
	reviewRepoSqlite := storage.NewReviewRepository(dbSQLite)
	reviewRepoPG := storage.NewReviewRepository(dbPostgres)

	reviews, err := reviewRepoSqlite.FindAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, review := range reviews {
		if err := reviewRepoPG.Store(review); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("success")
}
