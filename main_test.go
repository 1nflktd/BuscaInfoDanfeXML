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

    "."
)

var a main.App

func TestMain(m *testing.M) {
    a = main.App{}
    a.Initialize(os.Getenv("APP_ROOT_FOLDER_PATH"))

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

func TestGetDanfesWithoutFolder(t *testing.T) {
    clearFolder()

    req, _ := http.NewRequest("GET", "/danfes", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetDanfesInvalidFolder(t *testing.T) {
    clearFolder()

    req, _ := http.NewRequest("GET", "/danfes/123", nil)
    response := executeRequest(req)

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

    req, _ := http.NewRequest("POST", "/file", &buf)
    req.Header.Set("Content-Type", w.FormDataContentType())
    response := executeRequest(req)

    checkResponseCode(t, http.StatusCreated, response.Code)
}
