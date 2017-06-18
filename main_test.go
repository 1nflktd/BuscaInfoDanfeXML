package main_test

import (
    "os"
    "testing"
    "log"
    "golang.org/x/sys/unix"
    "net/http"
    "net/http/httptest"

    "."
)

var a main.App

func TestMain(m *testing.M) {
    a = main.App{}
    a.Initialize(os.Getenv("APP_ROOT_FOLDER"))

    ensurePathExistsAndWritable()

    code := m.Run()

    clearFolder()

    os.Exit(code)
}

func ensurePathExistsAndWritable() {
    if err := unix.Access(a.RootFolder, unix.W_OK); err != nil {
        log.Fatal(err)
    }
}

func clearFolder() {
    os.RemoveAll(a.RootFolder)
    os.MkdirAll(a.RootFolder, 0755)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
    rr := httptest.NewRecorder()
    a.Router.ServeHTTP(rr, req)

    return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}

func TestRequestWithoutFolder(t *testing.T) {
    clearFolder()

    req, _ := http.NewRequest("GET", "/danfes", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestRequestInvalidFolder(t *testing.T) {
    clearFolder()

    req, _ := http.NewRequest("GET", "/danfes/123", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}
