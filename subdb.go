package subdb

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	endPoint        = "http://api.thesubdb.com"
	sandboxEndPoint = "http://sandbox.thesubdb.com"
)

const (
	minFileSize = 2 * (64 * 1024)
)

var (
	// ErrNoUserAgent no user agent has been set
	ErrNoUserAgent = errors.New("no user agent set required to access SubDB API")
	// ErrNoSubtitle no subtitle found on the server
	ErrNoSubtitle = errors.New("no subtitle found for the requested hash on the server")
	// ErrSmallFileSize the file is to small to be hashed
	ErrSmallFileSize = errors.New("the supplied file is too small")
	// ErrDuplicated subtitle file already exists database
	ErrDuplicated = errors.New("subtitle file already exists in database")
	// ErrInvalidMediaType subtitle file is not supported
	ErrInvalidMediaType = errors.New("subtitle file is not supported by SubDB")
)

// API instance
type API struct {
	userAgent string
}

// SetUserAgent sets the user agent when accessing the API
func (a *API) SetUserAgent(clientName, clientVersion, clientURL string) {
	a.userAgent = fmt.Sprintf("SubDB/1.0 (%s/%s; %s)", clientName, clientVersion, clientURL)
}

// Languages lists the languages of all subtitles stored on the SubDB database
func (a *API) Languages() ([]string, error) {
	body, err := getRequest("action=languages", a.userAgent)
	if err != nil {
		return nil, err
	}
	return strings.Split(body, ","), nil
}

// Search list the available subtitle languages for a given file
func (a *API) Search(fileName string) ([]string, error) {
	hash, err := hashHelper(fileName)
	if err != nil {
		return nil, err
	}
	body, err := getRequest("action=search&hash="+hash+"&versions", a.userAgent)
	if err != nil {
		return nil, err
	}
	return strings.Split(body, ","), nil
}

// Download a subtitle for a given file in the provided language
func (a *API) Download(fileName, langCode string) (string, error) {
	hash, err := hashHelper(fileName)
	if err != nil {
		return "", err
	}
	body, err := getRequest("action=download&hash="+hash+"&language="+strings.ToLower(langCode), a.userAgent)
	if err != nil {
		return "", err
	}
	return body, nil
}

// Upload a subtitle file for a given file name
func (a *API) Upload(fileName, subtitlefile string) error {
	hash, err := hashHelper(fileName)
	if err != nil {
		return err
	}
	return postRequest(hash, subtitlefile, a.userAgent)
}

// hash hashes the first and last 64kb of a file
func hash(rs io.ReadSeeker) (string, error) {
	const readSize int64 = 64 * 1024
	first := make([]byte, readSize)
	last := make([]byte, readSize)
	if _, err := rs.Read(first); err != nil {
		return "", err
	}
	if _, err := rs.Seek(-readSize, os.SEEK_END); err != nil {
		return "", err
	}
	if _, err := rs.Read(last); err != nil {
		return "", err
	}
	var b bytes.Buffer
	b.Write(first)
	b.Write(last)
	h := md5.New()
	if _, err := io.Copy(h, &b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// hashHelper opens a file and runs the function 'hash' over it
func hashHelper(fileName string) (string, error) {
	fileInfo, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return "", err
	}
	if fileInfo.Size() < minFileSize {
		return "", ErrSmallFileSize
	}
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash, err := hash(f)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// getRequest performs a HTTP GET request
func getRequest(params string, userAgent string) (string, error) {
	if userAgent == "" {
		return "", ErrNoUserAgent
	}
	req, err := http.NewRequest("GET", sandboxEndPoint+"/?"+params, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Close = true
	c := &http.Client{Timeout: 5 * time.Second}
	resp, err := c.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", ErrNoSubtitle
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(http.StatusText(resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// postRequest performs a HTTP POST to upload a single subtitle
func postRequest(hash, subtitleFile, userAgent string) error {
	file, err := os.Open(subtitleFile)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(subtitleFile))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	_ = writer.WriteField("hash", hash)
	err = writer.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endPoint+"/?action=upload", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)
	req.Close = true
	c := &http.Client{}
	resp, err := c.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusForbidden:
		return ErrDuplicated
	case http.StatusUnsupportedMediaType:
		return ErrInvalidMediaType
	}
	return err
}
