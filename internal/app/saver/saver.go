package saver

import (
	"encoding/json"
	"os"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
)

type Saver struct {
	ds *diskSaver
}

func New(c *config.Config) *Saver {
	return &Saver{
		ds: newDiskSaver(c.FileStorage),
	}
}

func (s *Saver) Save(data *config.ShortURL) error {
	if s.ds != nil {
		return s.ds.Save(data)
	}
	return nil
}

func (s *Saver) Load(data map[string]*config.ShortURL) error {
	if s.ds != nil {
		return s.ds.Load(data)
	}
	return nil
}

type diskSaver struct {
	filename string
	file     *os.File
}

func newDiskSaver(filename string) *diskSaver {
	if filename == "" {
		return nil
	}
	return &diskSaver{
		filename: filename,
	}
}

func (ds *diskSaver) openFile() error {
	file, err := os.OpenFile(ds.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	ds.file = file
	return nil
}

func (ds *diskSaver) closeFile() {
	ds.file.Close()
}

func (ds *diskSaver) Save(data *config.ShortURL) error {
	if err := ds.openFile(); err != nil {
		return err
	}
	defer ds.closeFile()
	encoder := json.NewEncoder(ds.file)
	if err := encoder.Encode(&data); err != nil {
		return err
	}
	return nil
}

func (ds *diskSaver) Load(data map[string]*config.ShortURL) error {
	if err := ds.openFile(); err != nil {
		return err
	}
	defer ds.closeFile()

	decoder := json.NewDecoder(ds.file)
	for decoder.More() {
		shortRec := &config.ShortURL{}
		if err := decoder.Decode(shortRec); err != nil {
			return err
		}
		data[shortRec.Short] = shortRec
	}
	return nil
}
