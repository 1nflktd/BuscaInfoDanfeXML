package main_test

import (
    "os"
    "testing"
    "log"
    "golang.org/x/sys/unix"
    "net/http"
    "net/http/httptest"
    "mime/multipart"
    "bytes"
    "io"
    "path/filepath"
    "encoding/gob"
    "fmt"
    "strings"

    "github.com/alexedwards/scs/session"
    "github.com/alexedwards/scs/engine/memstore"
    "github.com/gorilla/mux"

    "."
)

var a main.App
var testEngine session.Engine
var testServeMux *mux.Router
var cookieName string

type User struct {
    ID int `storm:"increment"`
    Folder string `storm:"unique"`
}

func TestMain(m *testing.M) {
    a = main.App{}
    a.Initialize(os.Getenv("APP_ROOT_FOLDER_PATH"))
    defer a.Close()

    gob.Register(User{})
    
    // initialize storage engine
    testEngine = memstore.New(0)

    testServeMux = mux.NewRouter()
    testServeMux.HandleFunc("/authenticate", a.Authenticate).Methods("GET")
    testServeMux.HandleFunc("/file", a.PostFile).Methods("POST")
    testServeMux.HandleFunc("/import/{id:[0-9]+}", a.ImportFile).Methods("GET")
    testServeMux.HandleFunc("/danfes/{folder}", a.GetDanfes).Methods("GET")

    cookieName = "scs.session.token"

    ensurePathExistsAndWritable()

    code := m.Run()

    clearFolder()

    os.Exit(code)
}

func ensurePathExistsAndWritable() {
    if err := unix.Access(a.RootFolderPath, unix.W_OK); err != nil {
        log.Fatal(err)
    }
}

func clearFolder() {
    os.RemoveAll(a.RootFolderPath)
    os.MkdirAll(a.RootFolderPath, 0755)
}

func executeRequest(req *http.Request, h http.Handler) *httptest.ResponseRecorder {
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}

func extractTokenFromCookie(c string) string {
    parts := strings.Split(c, ";")
    return strings.TrimPrefix(parts[0], fmt.Sprintf("%s=", cookieName))
}

func TestAuthenticate(t *testing.T) {
    e := testEngine
    m := session.Manage(e)
    h := m(testServeMux)

    req, _ := http.NewRequest("GET", "/authenticate", nil)
    response := executeRequest(req, h)

    cookie := response.Header().Get("Set-Cookie")
    if cookie == "" {
        t.Errorf("Error authenticating, fail to generate cookie\n")
    }

    checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetDanfesWithoutFolder(t *testing.T) {
    clearFolder()

    e := testEngine
    m := session.Manage(e)
    h := m(testServeMux)
    req, _ := http.NewRequest("GET", "/danfes", nil)
    response := executeRequest(req, h)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetDanfesInvalidFolder(t *testing.T) {
    clearFolder()

    e := testEngine
    m := session.Manage(e)
    h := m(testServeMux)
    req, _ := http.NewRequest("GET", "/danfes/123", nil)
    response := executeRequest(req, h)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestPostFile(t *testing.T) {
    clearFolder()

    file := "/home/henrique/Documentos/BuscaInfoDanfeXML/testFiles/xml_nfe_2017-06-16-21-37-50.zip"

    var buf bytes.Buffer
    w := multipart.NewWriter(&buf)
    f, err := os.Open(file)
    if err != nil {
        t.Errorf("Error opening file: %v\n", err.Error()) 
    }
    defer f.Close()

    fw, err := w.CreateFormFile("file", filepath.Base(file))
    if err != nil {
        t.Errorf("Error creating form file: %v\n", err.Error())
    }

    if _, err = io.Copy(fw, f); err != nil {
        t.Errorf("Error copying file: %v\n", err.Error()) 
    }
    w.Close()

    e := testEngine
    m := session.Manage(e)
    h := m(testServeMux)
    req, _ := http.NewRequest("POST", "/file", &buf)
    req.Header.Set("Content-Type", w.FormDataContentType())
    response := executeRequest(req, h)

    checkResponseCode(t, http.StatusCreated, response.Code)
}
