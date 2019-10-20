package utils

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/proxy"
)

// GetHTTPClient - create HTTP client with proxy
func GetHTTPClient() *http.Client {
	if !Configuration.Proxy.Enabled {
		return &http.Client{}
	}

	dialer, err := proxy.SOCKS5("tcp", Configuration.Proxy.IP+":"+strconv.Itoa(Configuration.Proxy.Port), &proxy.Auth{User: Configuration.Proxy.User, Password: Configuration.Proxy.Password},
		proxy.Direct)

	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
		return nil
	}

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial

	return httpClient
}
