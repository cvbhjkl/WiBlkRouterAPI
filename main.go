package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"runtime"
	"github.com/PuerkitoBio/goquery"
)

const (
	NETWORK = "tcp"
	RADDR   = "127.0.0.1:8080"
)

type Console struct {
	client         *http.Client
	username       string
	password       string
	routerName     string
	routerPassword string
}

func (c *Console) Login() error {
	c.client = &http.Client{}
	url := "http://192.168.218.1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(c.username, c.password)
	rep, err := c.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(rep.Body)
	rep.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", data)
	return err
}

func (c *Console) BasicSettings() error {
	url := "http://192.168.218.1/internet/basic_settings.shtml"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(c.username, c.password)
	rep, err := c.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	doc,err:=goquery.NewDocumentFromReader(rep.Body)
	defer rep.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("table").Each(func(i int,s *goquery.Selection){
		s.Find("td").Each(func(i1 int,s1 *goquery.Selection){
			s2,exist:=s1.Find("#ssidname").Attr("value")
			if exist{
				c.routerName=s2
				log.Println("find router name: "+c.routerName)
			}
			s2,exist=s1.Find("#ssidpassword").Attr("value")
			if exist{
				c.routerPassword=s2
				log.Println("find router password: "+c.routerPassword)
			}
		})
	})
	return err
}
func ShowWiFiConfig() {
	//fmt.Println(runtime.GOARCH)
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "wlan", "show", "profiles")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out:\n%s\n", string(out))
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		log.Printf("combined out:\n%s\n", string(out))
	case "darwin":
		cmd := exec.Command("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-s")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out:\n%s\n", string(out))
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		log.Printf("combined out:\n%s\n", string(out))
	}
}
func ConnectToWiFi(rName string,rPassword string) {
	switch runtime.GOOS {
	case "windows":
		//TODO
		cmd := exec.Command("netsh", "wlan", "show", "profiles")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out:\n%s\n", string(out))
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		log.Printf("combined out:\n%s\n", string(out))
	case "darwin":
		cmd := exec.Command("networksetup", "-setairportnetwork" ,"en0", rName,rPassword)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("combined out:\n%s\n", string(out))
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		log.Printf("combined out:\n%s\n", string(out))
	}
}
func socketHandler(conn net.Conn) {
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(string(buf[:n]))
		request := string(buf[:n])
		rArr := strings.Split(request, " ")
		switch rArr[0] {
		case "request":
			fmt.Println("new request from " + rArr[2] + " asking for " + rArr[1])
			fmt.Println("please choose accept or refuse")
			i := <-cAction
			if i == 1 {
				conn.Write([]byte("accept" + " " + rArr[1] + " " + rArr[2]))
			} else {
				conn.Write([]byte("refuse" + " " + rArr[1] + " " + rArr[2]))
			}
		case "accept":
			fmt.Println("accept with " + rArr[1])
			ConnectToWiFi(rArr[1],rArr[2])
		case "refuse":
			fmt.Println("refused")
		}
	}
}

var console Console
var command string
var buf []byte = make([]byte, 4096)
var cAction chan int = make(chan int)

func main() {
	conn, err := net.DialTimeout(NETWORK, RADDR, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
	//conn.Write([]byte("hello socket\n"))
	log.Println("connect to server")
	go socketHandler(conn)
	for {
		fmt.Scanf("%s", &command)
		switch command {
		case "login":
			fmt.Scanf("%s %s", &console.username, &console.password)
			err := console.Login()
			if err != nil {
				log.Print(err)
				continue
			}
			console.BasicSettings()
			log.Println("log in success")
			log.Println("rName: "+console.routerName)
			log.Println("rPassword: "+console.routerPassword)
		case "register":
			conn.Write([]byte("register" + " " + console.routerName + " " + console.routerPassword))
		case "accept":
			cAction <- 1
		case "refuse":
			cAction <- 0
		case "request":
			var rName string
			fmt.Scanf("%s", &rName)
			conn.Write([]byte("request" + " " + rName))
		default:
			log.Println("not a legal command")
		}
	}
}
