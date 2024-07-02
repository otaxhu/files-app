package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/otaxhu/files-app/api"
	"github.com/otaxhu/files-app/repository"
	"github.com/otaxhu/files-app/service"
)

func main() {
	log.SetFlags(0)
	dir := flag.String("dir", "", "directory where the files are gonna be saved")
	dbFile := flag.String("db", "", "sqlite3 database file")
	reset := flag.Bool("reset", false, "if true, deletes 'dir' and 'db' and creates them again")
	frontendDir := flag.String("frontend", "", "directory where the frontend is")

	frontendPort := flag.Int("frontend-port", 0, "port where is gonna be served the frontend")
	backendPort := flag.Int("backend-port", 0, "port where is gonna be served the backend")

	flag.Parse()

	if *dir == "" || *dbFile == "" || *frontendDir == "" {
		log.Fatal("--dir, --frontend and --db options are required to be set")
	}

	dbAbs, err := filepath.Abs(*dbFile)
	if err != nil {
		log.Fatal(err)
	}

	dirAbs, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	frontendAbs, err := filepath.Abs(*frontendDir)
	if err != nil {
		log.Fatal(err)
	}

	if *reset {
		os.RemoveAll(dirAbs)
		os.RemoveAll(dbAbs)
	}

	if err := os.MkdirAll(dirAbs, 0777); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", dbAbs)
	if err != nil {
		log.Fatal(err)
	}

	fileRepo, err := repository.NewFileRepository(db, dirAbs)
	if err != nil {
		log.Fatal(err)
	}

	err = fileRepo.InitTables()
	if err != nil {
		log.Fatal(err)
	}

	fileServ := service.NewFileService(fileRepo)

	frontendFS := os.DirFS(frontendAbs)

	go api.StartFrontend(":"+strconv.Itoa(*frontendPort), frontendFS)
	api.Start(":"+strconv.Itoa(*backendPort), fileServ)
}
