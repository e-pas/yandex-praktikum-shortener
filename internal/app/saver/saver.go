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
	encoder  *json.Encoder
	decoder  *json.Decoder
}

func newDiskSaver(filename string) *diskSaver {
	if filename == "" {
		return nil
	}
	return &diskSaver{
		filename: filename,
	}
}

func (d *diskSaver) open() error {
	file, err := os.OpenFile(d.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	d.file = file
	return nil
}

func (d *diskSaver) close() {
	d.file.Close()
}

func (d *diskSaver) Save(data *config.ShortURL) error {
	if err := d.open(); err != nil {
		return err
	}
	defer d.close()
	encoder := json.NewEncoder(d.file)
	if err := encoder.Encode(&data); err != nil {
		return err
	}
	return nil
}

func (d *diskSaver) Load(data map[string]*config.ShortURL) error {
	if err := d.open(); err != nil {
		return err
	}
	defer d.close()

	decoder := json.NewDecoder(d.file)
	for decoder.More() {
		shortRec := &config.ShortURL{}
		if err := decoder.Decode(shortRec); err != nil {
			return err
		}
		data[shortRec.Short] = shortRec
	}
	return nil
}
