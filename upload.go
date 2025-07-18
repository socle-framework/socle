package socle

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"
	"github.com/socle-framework/filesystems"
)

func (s *Socle) UploadFile(r *http.Request, destination, field string, fs filesystems.FS) error {
	fileName, err := s.getFileToUpload(r, field)
	if err != nil {
		s.Log.ErrorLog.Println(err)
		return err
	}

	if fs != nil {
		err = fs.Put(fileName, destination)
		if err != nil {
			s.Log.ErrorLog.Println(err)
			return err
		}
	} else {
		err = os.Rename(fileName, fmt.Sprintf("%s/%s", destination, path.Base(fileName)))
		if err != nil {
			s.Log.ErrorLog.Println(err)
			return err
		}
	}

	defer func() {
		_ = os.Remove(fileName)
	}()

	return nil
}

func (s *Socle) getFileToUpload(r *http.Request, fieldName string) (string, error) {
	_ = r.ParseMultipartForm(s.env.uploads.maxUploadSize)

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	// go back to start of file
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}

	if !inSlice(s.env.uploads.allowedMimeTypes, mimeType.String()) {
		return "", errors.New("invalid file type uploaded")
	}

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
