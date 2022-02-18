package pinata

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"
	"time"
)

const (
	pinFileURL = "https://api.pinata.cloud/pinning/pinFileToIPFS"
)

type Pinata struct {
	Bearer string
}

type Response struct {
	IpfsHash  string `json:"IpfsHash"`
	PinSize   int64  `json:"PinSize"`
	Timestamp string `json:"Timestamp"`
}

type Error struct {
	Error struct {
		Reason  string `json:"reason"`
		Details string `json:"details"`
	} `json:"error"`
}

func (p *Pinata) SetKeys() {
	p.Bearer, _ = os.LookupEnv("PINATA_APIKEY")
}

//PinFile - upload file to IPFS
func (p *Pinata) PinFile(filepath string) (Response, error) {
	p.SetKeys()
	file, err := os.Open(filepath)

	if err != nil {
		return Response{}, err
	}
	defer func() { _ = file.Close() }()

	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer func() { _ = w.Close() }()
		defer func() { _ = m.Close() }()

		part, err := m.CreateFormFile("file", fp.Base(file.Name()))

		if err != nil {
			return
		}

		if _, err = io.Copy(part, file); err != nil {
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, pinFileURL, r)

	if err != nil {
		return Response{}, err
	}

	req.Header.Add("Content-Type", m.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+p.Bearer)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	if strings.Contains(string(data), "error") {
		var error Error
		if err := json.Unmarshal(data, &error); err != nil {
			return Response{}, err
		}

		return Response{}, errors.New(error.Error.Details)
	}

	var ret Response
	if err := json.Unmarshal(data, &ret); err != nil {
		return Response{}, err
	}

	return ret, nil

}
