package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var tikaPort int
var tikaAddress string
var bindPort int
var sleepTime time.Duration

func init() {
	// initialize global variables
	sleepTime = time.Duration(1) * time.Second

	flag.IntVar(&bindPort, "port", 9875, "The port to listen on")
	flag.IntVar(&tikaPort, "tika-port", 9876, "The port of the tika app")
	flag.StringVar(&tikaAddress, "tika-address", "localhost", "The address of the tika app")
	flag.Parse()
}

func main() {
	http.HandleFunc("/tika", call)
	http.HandleFunc("/", index)

	log.Printf("Listening on http://127.0.0.1:%d/", bindPort)
	http.ListenAndServe(fmt.Sprintf(":%d", bindPort), nil)
}

func resolveTikaAddr() (*net.TCPAddr, error) {
	var err error
	var tcpAddr *net.TCPAddr

	for i := 0; i < 10; i++ {
		tcpAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", tikaAddress, tikaPort))
		if err == nil {
			return tcpAddr, nil
		}
		time.Sleep(sleepTime)
	}
	return nil, err
}

func getTikaConn(tcpAddr *net.TCPAddr) (*net.TCPConn, error) {
	var err error
	var conn *net.TCPConn

	for i := 0; i < 10; i++ {
		conn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err == nil {
			return conn, nil
		}
		time.Sleep(sleepTime)
	}
	return nil, err
}

func readerToTika(reader io.Reader) (io.ReadCloser, error) {
	tcpAddr, err := resolveTikaAddr()
	if err != nil {
		log.Println("Error resolving tika address:", err)
		return nil, err
	}

	conn, err := getTikaConn(tcpAddr)
	if err != nil {
		log.Println("Error connecting to tika:", err)
		return nil, err
	}
	io.Copy(conn, reader)
	conn.CloseWrite()

	deadline := time.Now().Add(time.Second * time.Duration(60))
	conn.SetReadDeadline(deadline)

	if err != nil {
		log.Println("Read from Tika failed:", err.Error())
		return nil, err
	}

	return conn, nil
}

func downloadPdf(url string) (*http.Response, error) {
	var err error
	var resp *http.Response

	for i := 0; i < 10; i++ {
		resp, err = http.Get(url)
		if err == nil {
			return resp, nil
		}
		time.Sleep(sleepTime)
	}
	return nil, err
}

func urlToTika(url string) (io.ReadCloser, error) {
	resp, err := downloadPdf(url)
	if err != nil {
		log.Println("Error downloading PDF")
		return nil, err
	}

	defer resp.Body.Close()

	out, err := readerToTika(resp.Body)

	if err != nil {
		log.Println("Connection error:", err)
		return nil, err
	}

	return out, nil
}

func call(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	pdf := req.Form.Get("pdf")

	if pdf == "" {
		w.Write([]byte("You must supply a 'pdf' GET parameter (a URL)\n"))
		return
	}

	out, err := urlToTika(pdf)
	if err == nil {
		defer out.Close()
		_, err := io.Copy(w, out)
		if err != nil {
			w.Write([]byte("Error\n"))
			w.Write([]byte(err.Error()))
		}
	} else {
		w.Write([]byte("Error\n"))
		w.Write([]byte(err.Error()))
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	name := "templates/index.html"

	content, err := os.Open(name)
	if err != nil {
		w.Write([]byte("Error opening the index file"))
		return
	}
	defer content.Close()

	http.ServeContent(w, req, name, time.Time{}, content)
}
