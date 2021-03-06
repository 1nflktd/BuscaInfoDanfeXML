package main

import (
    "log"
    "net/http"
    "strconv"
    "encoding/gob"

    "github.com/gorilla/mux"
    "github.com/alexedwards/scs/engine/memstore"
    "github.com/alexedwards/scs/session"
    "github.com/asdine/storm"
)

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

func (a *App) initializeRoutes() {
    a.Router.HandleFunc("/authenticate", a.Authenticate).Methods("GET")
    a.Router.HandleFunc("/file", a.PostFile).Methods("POST")
    a.Router.HandleFunc("/import/{id:[0-9]+}", a.ImportFile).Methods("GET")
    a.Router.HandleFunc("/danfes", a.GetDanfes).Methods("GET")
}

func (a *App) Authenticate(w http.ResponseWriter, r *http.Request) {
	user := User{}

    // save user to database
    if err := a.DB.Save(&user); err != nil {
        log.Println(err.Error())
        error500(w)
    }

    // save user to session
    err := session.PutObject(r, "user", &user)
    if err != nil {
        log.Println(err.Error())
        error500(w)
    }

    respondWithJSON(w, http.StatusOK, nil);
}

func (a *App) PostFile(w http.ResponseWriter, r *http.Request) {
    // check if user authenticated and get user
    user, errAuth := getUserRequest(r)
    if errAuth != nil {
        log.Println("postFile, errAuth: " + errAuth.Error())
        respondWithError(w, http.StatusForbidden,  "User not authenticated")
    }

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
    	log.Println("postFile, errRemove: " + errRemove.Error())
        respondWithError(w, http.StatusInternalServerError, "Error removing file")
        return
    }

    newUser := &User{ID: user.ID, FolderName: folder.Name}

    a.DB.Update(newUser)

    errUpdateReq := updateRequestHandler(r, newUser)
    if errUpdateReq != nil {
        log.Println("postFile, errUpdateReq: " + errUpdateReq.Error())
        respondWithError(w, http.StatusInternalServerError, "Session error")
        return
    }

    respondWithJSON(w, http.StatusCreated, retMapUnzipFile)
}

func (a *App) ImportFile(w http.ResponseWriter, r *http.Request) {
    user, errAuth := getUserRequest(r)
    if errAuth != nil {
        log.Println("importFile, errAuth: " + errAuth.Error())
        respondWithError(w, http.StatusForbidden,  "User not authenticated")
        return
    }

	vars := mux.Vars(r)
	id, errConv := strconv.Atoi(vars["id"])
	if errConv != nil {
		log.Println("importFile, errConv: " + errConv.Error())
        respondWithError(w, http.StatusBadRequest, "Invalid parameter")
        return
	}

	// import file to db
    folder := &Folder{Name: user.FolderName, RootFolderPath: a.RootFolderPath}
    folder.genPath()
    folderFile := &FolderFile{FileNumber: id, Folder: folder}
    folderFile.Initialize()
    danfe, errDanfe := getDanfe(folderFile.Path)
    if errDanfe != nil {
        log.Println("importFile, errDanfe: " + errDanfe.Error())
        respondWithError(w, http.StatusBadRequest, "Invalid danfe file")
        return
    }

    danfe.UserID = user.ID
    if errSave := a.DB.Save(&danfe); errSave != nil {
        log.Println("importFile, errSave: " + errSave.Error())
        error500(w)
        return
    }

    respondWithJSON(w, http.StatusOK, nil)
}

func (a *App) GetDanfes(w http.ResponseWriter, r *http.Request) {
    user, errAuth := getUserRequest(r)
    if errAuth != nil {
        log.Println("importFile, errAuth: " + errAuth.Error())
        respondWithError(w, http.StatusForbidden,  "User not authenticated")
        return
    }

	// filter will be an array, ie. nome=produto 1&cfop=x505
	/*
    if errParse := r.ParseForm(); errParse != nil {
        log.Println("getDanfes, errParse: " + errParse.Error())
        error500(w)
        return
	}

	filter := make(map[string]string)
	for key, _ := range r.Form {
		filter[string(key)] = r.Form.Get(key)
	}
    */

    var danfes []Danfe
    if errFind := a.DB.Find("UserID", user.ID, &danfes); errFind != nil {
        log.Println("getDanfes, errFind: " + errFind.Error())
        error500(w)
        return
    }

    respondWithJSON(w, http.StatusOK, danfes)
}

func sessionError(w http.ResponseWriter, r *http.Request, err error) {
    log.Println(err.Error())
    error500(w)
}

func error500(w http.ResponseWriter) {
    respondWithError(w, http.StatusInternalServerError, "Sorry the application encountered an error")
}

func updateRequestHandler(r *http.Request, user *User) error {
    // update session token
    err := session.RegenerateToken(r)
    if err != nil {
        log.Println("updateRequestHandler, err 1: " + err.Error())
        return err
    }

    // update user
    err = session.PutObject(r, "user", user)
    if err != nil {
        log.Println("updateRequestHandler, err 2: " + err.Error())
        return err
    }

    return nil
}

func getUserRequest(r *http.Request) (*User, error) {
    user := &User{}
    err := session.GetObject(r, "user", user)
    return user, err
}
