package core

import (
	"bytes"
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
	query := r.URL.Query()
	query.Del("path")
	u.RawQuery = query.Encode()

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

	resProxyBody := io.NopCloser(bytes.NewBuffer(data))
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
