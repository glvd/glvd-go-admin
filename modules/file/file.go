// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package file

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/glvd/go-admin/plugins/admin/modules"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// Uploader is a file uploader which contains the method Upload.
type Uploader interface {
	Upload(*multipart.Form) error
}

// UploaderGenerator is a function return an Uploader.
type UploaderGenerator func() Uploader

var uploaderList = map[string]UploaderGenerator{
	"local": GetLocalFileUploader,
}

var upMu sync.Mutex

// AddUploader makes a uploader generator available by the provided theme name.
// If Add is called twice with the same name or if uploader is nil,
// it panics.
func AddUploader(name string, up UploaderGenerator) {
	upMu.Lock()
	defer upMu.Unlock()
	if up == nil {
		panic("uploader generator is nil")
	}
	if _, dup := uploaderList[name]; dup {
		panic("add uploader generator twice " + name)
	}
	uploaderList[name] = up
}

// GetFileEngine return the Uploader of given name.
func GetFileEngine(name string) Uploader {
	if up, ok := uploaderList[name]; ok {
		return up()
	}
	panic("wrong uploader name")
}

// UploadFun is a function to process the uploading logic.
type UploadFun func(*multipart.FileHeader, string) (string, error)

// Upload receive the return value of given UploadFun and put them into the form.
func Upload(c UploadFun, form *multipart.Form) error {
	var (
		suffix   string
		filename string
	)

	for k := range form.File {
		fileObj := form.File[k][0]
		suffix = path.Ext(fileObj.Filename)
		filename = modules.Uuid() + suffix

		pathStr, err := c(fileObj, filename)

		if err != nil {
			return err
		}
		form.Value["_filename_"+k] = []string{fileObj.Filename}
		form.Value["_filesize_"+k] = []string{humanize.Bytes(uint64(fileObj.Size))}
		form.Value[k] = []string{pathStr}
	}

	return nil
}

// SaveMultipartFile used in a local Uploader which help to save file in the local path.
func SaveMultipartFile(fh *multipart.FileHeader, path string) (err error) {
	var f multipart.File
	f, err = fh.Open()
	closed := false
	if err != nil {
		return err
	}
	defer func() {
		if !closed {
			if err2 := f.Close(); err2 != nil {
				err = err2
			}
		}
	}()

	if ff, ok := f.(*os.File); ok {
		err = f.Close()
		closed = true
		dir, name := filepath.Split(path)
		return moveFile(ff.Name(), dir, name, true)
	}

	ff, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := ff.Close(); err2 != nil {
			err = err2
		}
	}()
	_, err = copyZeroAlloc(ff, f)
	return err
}

func copyZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := io.CopyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}

var copyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func moveFile(sourcePath, toPath, destFile string, remove bool) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %w", err)
	}
	dest := filepath.Join(toPath, destFile)
	info, err := os.Stat(dest)
	if !os.IsNotExist(err) {
		if remove {
			sourceInfo, err := inputFile.Stat()
			if err != nil {
				return err
			}
			if info.Size() == sourceInfo.Size() {
				return os.Remove(sourcePath)
			}
		}
		return nil
	}

	err = os.Rename(sourcePath, dest)
	if err == nil {
		return nil
	} else {
		//ignore error
	}
	outputFile, err := os.Create(dest)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
