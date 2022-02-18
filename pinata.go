package pinata

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	fp "path/filepath"
	"time"
)

const (
	pinFileURL = "https://api.pinata.cloud/pinning/pinFileToIPFS"
)

type Pinata struct {
	apiKey string
	secret string
}

type Response struct {
	IpfsHash  string `json:"IpfsHash"`
	PinSize   string `json:"PinSize"`
	Timestamp int64  `json:"Timestamp"`
}

func (p *Pinata) SetKeys() {
	p.apiKey, _ = os.LookupEnv("PINATA_APIKEY")
	p.secret, _ = os.LookupEnv("PINATA_APISECRET")
}

//PinFile - upload file to IPFS
func (p *Pinata) PinFile(filepath string) (Response, error) {
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
	req.Header.Add("pinata_secret_api_key", p.apiKey)
	req.Header.Add("pinata_api_key", p.secret)

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

	var ret Response
	if err := json.Unmarshal(data, &ret); err != nil {
		return Response{}, err
	}

	return ret, nil

}
