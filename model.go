package main

import (
	"encoding/xml"
	"path/filepath"
	"time"
	"os"
//	"errors"
)

type Danfe struct {
	//XMLName xml.Name `xml:"nfeProc"`
	DataHoraEmissao time.Time `xml:"NFe>infNFe>ide>dhEmi" json:"dataHoraEmissao"`
	NomeDestinatario string `xml:"NFe>infNFe>dest>xNome" json:"nomeDestinatario"`
	ChaveNFe string `xml:"protNFe>infProt>chNFe" json:"chaveNFe"`
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
