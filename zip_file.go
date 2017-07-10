package main

import (
	"archive/zip"
    "path/filepath"
    "os"
    "io"
    "errors"
)

type ZipFile struct {
    Name string
    Path string
    Folder *Folder
}

func (z *ZipFile) Initialize() {
    z.Path = filepath.Join(z.Folder.Path, z.Name)
}

func (z *ZipFile) remove() error {
    return os.Remove(z.Path)
}

func (z *ZipFile) unzip() (retMap map[string]interface{}, retErr error) {
    r, err := zip.OpenReader(z.Path)
    if err != nil {
        return nil, err
    }
    defer func() {
        if err := r.Close(); err != nil {
            retMap, retErr = nil, err
        }
    }()

    os.MkdirAll(z.Folder.Path, 0755)

    filesFound := 0
    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) (errExtract error) {
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                errExtract = err
            }
        }()

        path := filepath.Join(z.Folder.Path, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            if filepath.Ext(path) != ".xml" {
                return errors.New("Not XML file")
            }

            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    errExtract = err
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }

            folderFile := &FolderFile{Folder: z.Folder, FileNumber: filesFound}
            folderFile.Initialize()
            errRename := folderFile.rename(path)
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
