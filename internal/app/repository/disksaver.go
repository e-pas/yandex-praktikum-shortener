package repository

import (
	"context"
	"encoding/json"
	"os"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/model"
)

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

func (ds *diskSaver) Save(ctx context.Context, data model.ShortURL) error {
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

func (ds *diskSaver) SaveBatch(ctx context.Context, data []*model.ShortURL) error {
	for _, rec := range data {
		if err := ds.Save(ctx, *rec); err != nil {
			return err
		}
	}
	return nil
}

func (ds *diskSaver) UpdateBatch(ctx context.Context, data []*model.ShortURL) error {
	return ds.SaveBatch(ctx, data)
}

func (ds *diskSaver) Load(ctx context.Context, data map[string]*model.ShortURL) error {
	if err := ds.openFile(); err != nil {
		return err
	}
	defer ds.closeFile()

	decoder := json.NewDecoder(ds.file)
	for decoder.More() {
		shortRec := &model.ShortURL{}
		if err := decoder.Decode(&shortRec); err != nil {
			return err
		}
		data[shortRec.Short] = shortRec
	}
	return nil
}

func (ds *diskSaver) Ping(ctx context.Context) error {
	if err := ds.openFile(); err != nil {
		return err
	}
	if err := ds.closeFile(); err != nil {
		return err
	}
	return nil
}

func (ds *diskSaver) openFile() error {
	file, err := os.OpenFile(ds.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	ds.file = file
	return nil
}

func (ds *diskSaver) closeFile() error {
	return ds.file.Close()
}
