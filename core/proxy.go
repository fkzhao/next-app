package core

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// DoProxy proxy http request
func DoProxy(w http.ResponseWriter, r *http.Request) {
	// create http client
	cli := &http.Client{}

	// read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Read body error:", err)
		// response code
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	u, err := url.Parse(os.Getenv("REMOTE_HTTP_SERVER"))
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}
	u.Path = r.URL.Path
	u.RawQuery = r.URL.RawQuery

	reqProxy, err := http.NewRequest(r.Method, u.String(), strings.NewReader(string(body)))
	if err != nil {
		log.Println("create request error:", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// rewrite request Header
	for k, v := range r.Header {
		reqProxy.Header.Set(k, v[0])
	}
	cookie := r.Header.Get("Cookie")
	if cookie != "" {
		cookie = "abuse_interstitial=" + u.Host

	} else {
		cookie = r.Header.Get("Cookie") + ";abuse_interstitial=" + u.Host
	}
	reqProxy.Header.Set("Cookie", cookie)

	// call remote api
	responseProxy, err := cli.Do(reqProxy)
	if err != nil {
		log.Println("call remote server error:", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(responseProxy.Body)

	// response Header
	for k, v := range responseProxy.Header {
		w.Header().Set(k, v[0])
	}

	// response body
	var data []byte

	//read response body
	data, err = io.ReadAll(responseProxy.Body)
	if err != nil {
		log.Println("response body read error:", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// write data to client
	var dataOutput []byte
	isGzipped := isGzipped(responseProxy.Header)
	if isGzipped {
		resProxyGzippedBody := io.NopCloser(bytes.NewBuffer(data))
		defer func(resProxyGzippedBody io.ReadCloser) {
			err := resProxyGzippedBody.Close()
			if err != nil {

			}
		}(resProxyGzippedBody) // delay close

		// gzip Reader
		gr, err := gzip.NewReader(resProxyGzippedBody)
		if err != nil {
			log.Println("create gzip reader error:", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer func(gr *gzip.Reader) {
			err := gr.Close()
			if err != nil {

			}
		}(gr)

		// read gzip data
		dataOutput, err = io.ReadAll(gr)
		if err != nil {
			log.Println("read gzip data error:", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	} else {
		dataOutput = data
	}

	resProxyBody := io.NopCloser(bytes.NewBuffer(dataOutput))
	defer func(resProxyBody io.ReadCloser) {
		err := resProxyBody.Close()
		if err != nil {

		}
	}(resProxyBody)

	w.WriteHeader(responseProxy.StatusCode)
	_, err = io.Copy(w, resProxyBody)
	if err != nil {
		log.Println("write response data to client failed:", err)
		return
	}
}

const headerContentEncoding = "Content-Encoding"
const encodingGzip = "gzip"

func isGzipped(header http.Header) bool {
	if header == nil {
		return false
	}

	contentEncoding := header.Get(headerContentEncoding)
	isGzipped := false
	if strings.Contains(contentEncoding, encodingGzip) {
		isGzipped = true
	}

	return isGzipped
}
