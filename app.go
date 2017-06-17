
package main

import (
    "log"
    "net/http"
    "os"
    "encoding/json"
    "io"
    "path/filepath"

    "github.com/gorilla/mux"
)

type App struct {
    Router *mux.Router
    RootFolder string
}

func (a *App) Initialize(rootFolder string) {
	a.RootFolder = rootFolder
    a.Router = mux.NewRouter()
    a.initializeRoutes()
}

func (a *App) Run(addr string) {
    log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
    a.Router.HandleFunc("/uploadFile", a.uploadFile).Methods("POST")
    a.Router.HandleFunc("/danfes", a.getDanfes).Methods("GET")
}

func (a *App) uploadFile(w http.ResponseWriter, r *http.Request) {
    file, header, errFormFile := r.FormFile("file")
    if errFormFile != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload: " + errFormFile.Error())
        return
    }
    defer file.Close()

    uniqueFolder := filepath.Join(a.RootFolder, a.getUniqueFolder())
    // Create a folder with unique name for this file
    if err := os.Mkdir(uniqueFolder, os.ModePerm); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    // unzip these file to that folder
	f, errOpenFile := os.OpenFile(filepath.Join(uniqueFolder, header.Filename), os.O_WRONLY | os.O_CREATE, 0666)
	if errOpenFile != nil {
        respondWithError(w, http.StatusInternalServerError, errOpenFile.Error())
	   return
	}
	defer f.Close()

	io.Copy(f, file)

    respondWithJSON(w, http.StatusCreated, nil)
}

func (a *App) getUniqueFolder() string {
	return "folder/"; // some way to retrieve a unique folder to the user
}

func (a *App) getDanfes(w http.ResponseWriter, r *http.Request) {
	// filter will be an array, ie. nome=produto 1&cfop=x505
	if err := r.ParseForm(); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
	}

	filter := make(map[string]string)
	for key, _ := range r.Form {
		filter[string(key)] = r.Form.Get(key)
	}

    folder := a.getUniqueFolder()

    danfes, err := getDanfes(a.RootFolder + folder, filter)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, danfes)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}
