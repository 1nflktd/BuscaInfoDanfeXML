package main

import (
    "path/filepath"
    "os"
    "io"
    "mime/multipart"
)

type Folder struct {
    Name string
    Path string
    RootFolderPath string
}

func (f *Folder) Initialize() {
    f.Name = generateUniqueString(32)
    f.Path = filepath.Join(f.RootFolderPath, f.Name)
}

func (f *Folder) upload(multiPartFile multipart.File, filename string) error {
    if err := os.Mkdir(f.Path, os.ModePerm); err != nil {
        return err
    }

    file, errOpenFile := os.OpenFile(filepath.Join(f.Path, filename), os.O_WRONLY | os.O_CREATE, 0666)
    if errOpenFile != nil {
        return errOpenFile
    }
    defer file.Close()

    io.Copy(file, multiPartFile)

    return nil
}

