
package main

import (
    "log"
    "net/http"
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

    folder, errUpload := uploadAndCreateFolder(file, header.Filename, a.RootFolder)
    if errUpload != nil {
        respondWithError(w, http.StatusBadRequest, errUpload.Error())
        return
    }

    filePath := filepath.Join(folder, header.Filename)
    errUnzip := unzipFile(filePath, folder)
    if errUnzip != nil {
        respondWithError(w, http.StatusBadRequest, errUnzip.Error())
        return
    }

    errRemove := removeFile(filePath)
    if errRemove != nil {
        respondWithError(w, http.StatusBadRequest, errRemove.Error())
        return    	
    }

    respondWithJSON(w, http.StatusCreated, nil)
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

    folder, errFolder := getUniqueString()
    if errFolder != nil {
        respondWithError(w, http.StatusInternalServerError, errFolder.Error())
        return
    }

    danfes, err := getDanfes(a.RootFolder + folder, filter)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, danfes)
}
