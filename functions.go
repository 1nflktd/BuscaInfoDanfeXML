package main

import (
	"crypto/rand"
	"encoding/base64"
	"archive/zip"
    "path/filepath"
    "os"
    "io"
    "io/ioutil"
    "mime/multipart"
    "errors"
    "strconv"
)

func getUniqueString() (string, error) {
	// from https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/
	b := make([]byte, 32)
	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), err
}

func uploadAndCreateFolder(file multipart.File, filename, rootFolder string) (string, error) {
    uniqueFolder, errFolder := getUniqueString()
    if errFolder != nil {
    	return "", errFolder
    }
    uniqueFolder = filepath.Join(rootFolder, uniqueFolder)
    // Create a folder with unique name for this file
    if err := os.Mkdir(uniqueFolder, os.ModePerm); err != nil {
    	return "", err
    }

	f, errOpenFile := os.OpenFile(filepath.Join(uniqueFolder, filename), os.O_WRONLY | os.O_CREATE, 0666)
	if errOpenFile != nil {
		return "", errOpenFile
	}
	defer f.Close()

	io.Copy(f, file)

	return uniqueFolder, nil
}

func encodeFileName(number int) string {
    return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number)))
}

func decodeFileName(hashBase64 string) ([]byte, error) {
    return base64.StdEncoding.DecodeString(hashBase64)
}

func unzipFile(src, dest string) (map[string]interface{}, error) {
    r, err := zip.OpenReader(src)
    if err != nil {
        return nil, err
    }
    defer func() {
        if err := r.Close(); err != nil {
            panic(err)
        }
    }()

    os.MkdirAll(dest, 0755)

    filesFound := 0
    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) error {
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        path := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            if filepath.Ext(path) != ".xml" {
                return errors.New("Not XML file")
            }

            dir := filepath.Dir(path)
            os.MkdirAll(dir, f.Mode())
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }

            // change name
            newName := encodeFileName(filesFound)
            newPath := filepath.Join(dir, newName + ".xml")
            errRename := os.Rename(path, newPath)
            if errRename != nil {
                return errRename
            }

            filesFound++
        }
        return nil
    }

    for _, f := range r.File {
        err := extractAndWriteFile(f)
        if err != nil {
            return nil, err
        }
    }

    ret := make(map[string]interface{})
    ret["total"] = filesFound
    return ret, nil
}

func removeFile(src string) error {
	return os.Remove(src)
}

func getFile(src string) ([]byte, error) {
	return ioutil.ReadFile(src)
}

func isDirectory(src string) (bool, error) {
	fileInfo, err := os.Stat(src)
	if err == nil {
	    return fileInfo.IsDir(), err
	}
	return false, err
}
