package main

import (
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/polygonledger/node/netio"
	"github.com/polygonledger/node/xutils"
)

const test_node_port_pub = 8888

func initserverpub() *TCPNode {
	// Start the new server

	testsrv, err := NewNode()
	testsrv.addr = ":" + strconv.Itoa(test_node_port_pub)

	if err != nil {
		log.Println("error starting TCP server")
		return testsrv
	} else {
		//log.Println("start ", testsrv)
	}

	go testsrv.RunTCP()
	go testsrv.HandleConnectTCP()

	return testsrv
}

func testclientpub() netio.Ntchan {
	time.Sleep(200 * time.Millisecond)
	addr := ":" + strconv.Itoa(test_node_port_pub)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		//t.Error("could not connect to server: ", err)
	}
	//t.Error("...")
	log.Println("connected")
	ntchan := netio.ConnNtchan(conn, "client", addr, true)
	//defer conn.Close()
	return ntchan

}

func TestServer_Pub(t *testing.T) {

	testsrv := initserverpub()
	defer testsrv.Close()

	ntchan := netio.ConnNtchanStub("server")

	time.Sleep(100 * time.Millisecond)

	go netio.PublishTime(ntchan)
	go netio.PubWriterLoop(ntchan)

	time.Sleep(100 * time.Millisecond)

	if xutils.IsEmpty(ntchan.Writer_queue, 1*time.Millisecond) {
		t.Error("Writer_queue empty")
	}

	// go SubLoop(ntchan)

	// reqs := "hello world"
	// n, err := netio.NetWrite(firstpeer.NTchan, reqs)

	// if err != nil {
	// 	t.Error("could not write to server:", err)
	// }

	// delimsize := 1
	// l := len([]byte(reqs)) + delimsize
	// if n != l {
	// 	t.Error("wrong bytes written ", l)
	// }

	// time.Sleep(100 * time.Millisecond)

	// //log.Println(clientNt.SrcName)

	// rmsg := <-clientNt.Reader_queue
	// if rmsg != reqs {
	// 	t.Error("different message on reader ", rmsg)
	// }
	// if isEmpty(clientNt.Reader_queue, 1*time.Second) {
	// 	t.Error("fail")
	//}

}
