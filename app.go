package main

import (
    "log"
    "net/http"
    "path/filepath"
    "strconv"
    "encoding/gob"

    "github.com/gorilla/mux"
    "github.com/alexedwards/scs/engine/memstore"
    "github.com/alexedwards/scs/session"
)

type User struct {
	Folder string
}

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
	gob.Register(User{})
	
	// initialize storage engine
	engine := memstore.New(0)

	// initialize session manager
	sessionManager := session.Manage(engine,
	    //session.IdleTimeout(30*time.Minute),
	    session.ErrorFunc(sessionError),
	)

    log.Fatal(http.ListenAndServe(addr, sessionManager(a.Router)))
}

func sessionError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(err.Error())
	respondWithError(w, http.StatusInternalServerError, "Sorry the application encountered an error")
}

func (a *App) initializeRoutes() {
    a.Router.HandleFunc("/authenticate", a.authenticate).Methods("GET")
    a.Router.HandleFunc("/file", a.postFile).Methods("POST")
    a.Router.HandleFunc("/import/{id:[0-9]+}", a.importFile).Methods("GET")
    a.Router.HandleFunc("/danfes/{folder}", a.getDanfes).Methods("GET")
}

func (a *App) authenticate(w http.ResponseWriter, r *http.Request) {
	user := &User{}
	err := session.PutObject(r, "user", user)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusInternalServerError, "Sorry the application encountered an error")
	}	
}

func (a *App) postFile(w http.ResponseWriter, r *http.Request) {
    file, header, errFormFile := r.FormFile("file")
    if errFormFile != nil {
    	log.Println("postFile, errFormFile: " + errFormFile.Error())
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    defer file.Close()

    folder, errUpload := uploadAndCreateFolder(file, header.Filename, a.RootFolder)
    if errUpload != nil {
    	log.Println("postFile, errUpload: " + errUpload.Error())
        respondWithError(w, http.StatusInternalServerError, "Error uploading file")
        return
    }

    filePath := filepath.Join(folder, header.Filename)
    retMapUnzipFile, errUnzip := unzipFile(filePath, folder)
    if errUnzip != nil {
    	log.Println("postFile, errUnzip: " + errUnzip.Error())
        respondWithError(w, http.StatusInternalServerError, "Error unziping file")
        return
    }

    errRemove := removeFile(filePath)
    if errRemove != nil {
    	log.Println("postFile, errUnzip: " + errRemove.Error())
        respondWithError(w, http.StatusInternalServerError, "Error removing file")
        return
    }

    respondWithJSON(w, http.StatusCreated, retMapUnzipFile)
}

func (a *App) importFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := strconv.Atoi(vars["id"]) // id
	if err != nil {
		log.Println("importFile: " + err.Error())
        respondWithError(w, http.StatusBadRequest, "Invalid parameter")
        return
	}

	// import file to db
}

func (a *App) getDanfes(w http.ResponseWriter, r *http.Request) {
	// filter will be an array, ie. nome=produto 1&cfop=x505
	vars := mux.Vars(r)
	folder := vars["folder"]
	if folder == "" {
        respondWithError(w, http.StatusBadRequest, "Folder not found in request.")
        return
	}

	if err := r.ParseForm(); err != nil {
        respondWithError(w, http.StatusInternalServerError, "Error parsing form: " + err.Error())
        return
	}

	filter := make(map[string]string)
	for key, _ := range r.Form {
		filter[string(key)] = r.Form.Get(key)
	}

	folderPath := filepath.Join(a.RootFolder, folder)

	isDir, err := isDirectory(folder)
	if err != nil || !isDir {
        respondWithError(w, http.StatusNotFound, "Folder not found")
        return
	}
	
    danfes, err := getDanfes(folderPath, filter)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Error getting danfes: " + err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, danfes)
}
