package main

import (
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	NETWORK string = "tcp"
	LADDR   string = "127.0.0.1:8080"
)

var mUid2Conn map[uint32]net.Conn = make(map[uint32]net.Conn)
var mRid2Password map[uint32]string = map[uint32]string{}
var mRid2Uid map[uint32]uint32 = make(map[uint32]uint32)
var mRname2Rid map[string]uint32 = make(map[string]uint32)
var nowRid uint32

func main() {
	nowUid := uint32(0)
	nowRid = 0
	listener, err := net.Listen(NETWORK, LADDR)
	if err != nil {
		log.Printf("listen to port error:%s", err)
		os.Exit(1)
	}
	log.Printf("now listen:err%s", err)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("connect error:err%s", err)
			os.Exit(1)
		}
		nowUid += 1
		mUid2Conn[nowUid] = conn
		go connHandle(conn, nowUid)
	}
}

func connHandle(conn net.Conn, uid uint32) {
	defer conn.Close()
	var buf []byte = make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("socket closed")
				break
			} else {
				log.Println("error:" + err.Error())
				break
			}
		}
		request := string(buf[:n])
		log.Println(request)
		rArr := strings.Split(request, " ")
		log.Println(rArr)
		switch rArr[0] {
		case "register":
			if len(rArr) < 3 {
				continue
			}
			nowRid += 1
			mRname2Rid[rArr[1]] = nowRid
			mRid2Uid[nowRid] = uid
			mRid2Password[nowRid] = rArr[2]
		case "request":
			if len(rArr) < 2 {
				continue
			}
			rid:=mRname2Rid[rArr[1]]
			log.Println(rid)
			uid:=mRid2Uid[rid]
			log.Println(uid)
			con:=mUid2Conn[uid]
			str1:="request" + " " + rArr[1] + " " + strconv.Itoa(int(uid))
			con.Write([]byte(str1))
		case "accept":
			if len(rArr) < 3 {
				continue
			}
			i, _ := strconv.Atoi(rArr[2])
			mUid2Conn[uint32(i)].Write([]byte("accept" + " " +rArr[1]+" "+ mRid2Password[mRname2Rid[rArr[1]]]))
		case "refuse":
			if len(rArr) < 3 {
				continue
			}
			i, _ := strconv.Atoi(rArr[2])
			mUid2Conn[uint32(i)].Write([]byte("refuse"))
		}
	}
}
