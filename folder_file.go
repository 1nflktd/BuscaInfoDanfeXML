package main

import (
	"encoding/base64"
    "path/filepath"
    "os"
    "strconv"
)

type FolderFile struct {
    Name string
    Path string
    FileNumber int
    Folder *Folder
}

func (f *FolderFile) Initialize() {
    f.setEncodedName(f.FileNumber)
    f.Path = filepath.Join(f.Folder.Path, f.Name + ".xml")
}

func (f *FolderFile) setEncodedName(number int) {
    f.Name = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number)))
}

func (f *FolderFile) getDecodedName() ([]byte, error) {
    return base64.StdEncoding.DecodeString(f.Name)
}

func (f *FolderFile) rename(path string) error {
    return os.Rename(path, f.Path)
}

