package main

import (
	"encoding/xml"
	"path/filepath"
	"time"
	"os"
)

type User struct {
    ID int `storm:"increment" json:"id"`
	FolderName string `storm:"unique" json:"folderName"`
}

type Danfe struct {
	ID int `storm:"increment" json:"id"`
	UserID int `storm:"index" json:"userId"`
	DataHoraEmissao time.Time `xml:"NFe>infNFe>ide>dhEmi" json:"dataHoraEmissao"`
	NomeDestinatario string `xml:"NFe>infNFe>dest>xNome" json:"nomeDestinatario"`
	ChaveNFe string `storm:"unique" xml:"protNFe>infProt>chNFe" json:"chaveNFe"`
}

func getDanfes(folder string, filters map[string]string) ([]Danfe, error) {
	var danfes []Danfe

	fVisit := func(path string, f os.FileInfo, err error) error {
		if isDir, err := isDirectory(path); isDir || err != nil {
			return nil
		}

		xmlDanfe, err := getFile(path)
		if err != nil {
			return err
		}

		danfe := &Danfe{}
		err = xml.Unmarshal(xmlDanfe, danfe)
		if err != nil {
			return err
		}

		danfes = append(danfes, *danfe)

		return nil
	}

	err := filepath.Walk(folder, fVisit)
	if err != nil {
		return nil, err
	}

	return danfes, nil
}

func getDanfe(path string) (Danfe, error) {
	xmlDanfe, err := getFile(path)
	if err != nil {
		return Danfe{}, err
	}

	danfe := Danfe{}
	err = xml.Unmarshal(xmlDanfe, &danfe)
	if err != nil {
		return Danfe{}, err
	}

	return danfe, nil
}
