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
    "github.com/asdine/storm"
)

type User struct {
    ID int `storm:"increment"`
	Folder string `storm:"unique"`
}

type App struct {
    Router *mux.Router
    DB *storm.DB
    RootFolderPath string    
}

func (a *App) Initialize(rootFolderPath string, dbName string) {
    var err error
    a.DB, err = storm.Open(dbName)
    if err != nil {
        log.Fatal(err)
    }
	a.RootFolderPath = rootFolderPath
    a.Router = mux.NewRouter()
    a.initializeRoutes()
}

func (a *App) Close() {
    a.DB.Close()
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
    a.Router.HandleFunc("/authenticate", a.Authenticate).Methods("GET")
    a.Router.HandleFunc("/file", a.PostFile).Methods("POST")
    a.Router.HandleFunc("/import/{id:[0-9]+}", a.ImportFile).Methods("GET")
    a.Router.HandleFunc("/danfes/{folder}", a.GetDanfes).Methods("GET")
}

func (a *App) Authenticate(w http.ResponseWriter, r *http.Request) {
	user := User{}

    // save user to database
    if err := a.DB.Save(&user); err != nil {
        log.Println(err.Error())
        respondWithError(w, http.StatusInternalServerError, "Sorry the application encountered an error")
    }
    
    // save user to session
    err := session.PutObject(r, "user", &user)
    if err != nil {
        log.Println(err.Error())
        respondWithError(w, http.StatusInternalServerError, "Sorry the application encountered an error")
    }
    
    respondWithJSON(w, http.StatusOK, nil);
}

func (a *App) PostFile(w http.ResponseWriter, r *http.Request) {
    file, header, errFormFile := r.FormFile("file")
    if errFormFile != nil {
    	log.Println("postFile, errFormFile: " + errFormFile.Error())
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    defer file.Close()

    folder := &Folder{RootFolderPath: a.RootFolderPath}
    folder.Initialize()
    errUpload := folder.upload(file, header.Filename)
    if errUpload != nil {
    	log.Println("postFile, errUpload: " + errUpload.Error())
        respondWithError(w, http.StatusInternalServerError, "Error uploading file")
        return
    }

    zipFile := &ZipFile{Folder: folder, Name: header.Filename}
    zipFile.Initialize()
    retMapUnzipFile, errUnzip := zipFile.unzip()
    if errUnzip != nil {
    	log.Println("postFile, errUnzip: " + errUnzip.Error())
        respondWithError(w, http.StatusInternalServerError, "Error unziping file")
        return
    }

    errRemove := zipFile.remove()
    if errRemove != nil {
    	log.Println("postFile, errUnzip: " + errRemove.Error())
        respondWithError(w, http.StatusInternalServerError, "Error removing file")
        return
    }

    respondWithJSON(w, http.StatusCreated, retMapUnzipFile)
}

func (a *App) ImportFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("importFile: " + err.Error())
        respondWithError(w, http.StatusBadRequest, "Invalid parameter")
        return
	}

	// import file to db
}

func (a *App) GetDanfes(w http.ResponseWriter, r *http.Request) {
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

	folderPath := filepath.Join(a.RootFolderPath, folder)

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
