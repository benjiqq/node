package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/polygonledger/node/block"
	"github.com/polygonledger/node/chain"
	"github.com/polygonledger/node/ntcl"
	"github.com/polygonledger/node/ntwk"
)

//simple node that runs standalone without peers

//var srv Server

const node_port = 8888

var nlog *log.Logger
var logfile_name = "node.log"

type Configuration struct {
	PeerAddresses []string
	NodePort      int
	WebPort       int
}

type TCPServer struct {
	Name          string
	addr          string
	server        net.Listener
	accepting     bool
	ConnectedChan chan net.Conn
	//TODO! list of peers
	Peers []ntcl.Peer
}

func (t *TCPServer) GetPeers() []ntcl.Peer {
	if &t.Peers == nil {
		return nil
	}
	return t.Peers
}

// start listening on tcp and handle connection through channels
func (t *TCPServer) Run() (err error) {

	log.Println("node listen on ", t.addr)
	t.server, err = net.Listen("tcp", t.addr)
	if err != nil {
		//return errors.Wrapf(err, "Unable to listen on port %s\n", t.addr)
	}
	//run forever and don't close
	//defer t.Close()

	for {
		t.accepting = true
		conn, err := t.server.Accept()
		if err != nil {
			err = errors.New("could not accept connection")
			break
		}
		if conn == nil {
			err = errors.New("could not create connection")
			break
		}

		log.Println("put on out ", conn)
		t.ConnectedChan <- conn

	}
	log.Println("end run")
	return
}

func (t *TCPServer) HandleDisconnect() {

}

//handle new connection
func (t *TCPServer) HandleConnect() {

	//TODO! hearbeart, check if peers are alive
	//TODO! handshake

	for {
		newpeerConn := <-t.ConnectedChan
		strRemoteAddr := newpeerConn.RemoteAddr().String()
		log.Println("accepted conn ", strRemoteAddr, t.accepting)
		log.Println("new peer ", newpeerConn)
		// log.Println("> ", t.Peers)
		// log.Println("# peers ", len(t.Peers))

		ntchan := ntcl.ConnNtchan(newpeerConn, "server", strRemoteAddr)

		p := ntcl.Peer{Address: strRemoteAddr, NodePort: node_port, NTchan: ntchan}
		t.Peers = append(t.Peers, p)

		go t.handleConnection(ntchan)
		//go ChannelPeerNetwork(conn, peer)
		//setupPeer(strRemoteAddr, node_port, conn)

		//conn.Close()

	}
}

func echohandler(ins string) string {
	resp := "Echo:" + ins
	return resp
}

func (t *TCPServer) handleConnection(ntchan ntcl.Ntchan) {
	//tr := 100 * time.Millisecond
	//defer ntchan.Conn.Close()
	log.Println("handleConnection")

	go ntcl.ReadLoop(ntchan)
	go ntcl.ReadProcessor(ntchan)
	go ntcl.WriteLoop(ntchan, 500*time.Millisecond)

	go func() {
		for {
			msg_string := <-ntchan.REQ_in
			log.Println("handle ", msg_string)
			msg := ntwk.ParseMessage(msg_string)
			//reply_string := "default"

			//"echo:" + msg_string
			var reply string
			var reply_msg string

			switch msg.Command {

			case ntwk.CMD_PING:
				reply = "PONG"
				reply_msg = ntwk.EncodeMsgString(ntwk.REP, reply, "")

			case ntwk.CMD_BALANCE:
				nlog.Println("Handle balance")

				dataBytes := msg.Data
				nlog.Println("data ", string(msg.Data), dataBytes)

				a := block.Account{AccountKey: string(msg.Data)}

				// var account block.Account

				// if err := json.Unmarshal(dataBytes, &account); err != nil {
				// 	panic(err)
				// }
				// nlog.Println("get balance for account ", account)

				balance := chain.Accounts[a]
				//s := strconv.Itoa(balance)
				// data, _ := json.Marshal(balance)
				data := strconv.Itoa(balance)
				reply_msg = ntwk.EncodeMsgString(ntwk.REP, ntwk.CMD_BALANCE, data)
				log.Println(">> ", reply)

				//rep_chan <- reply

			}

			ntchan.Writer_queue <- reply_msg
		}
	}()
}

//HTTP
func LoadContent() string {
	content := ""

	// content += fmt.Sprintf("<h2>Peers</h2>Peers: %d<br>", len(peers))
	// for i := 0; i < len(peers); i++ {
	// 	content += fmt.Sprintf("peer ip address: %s<br>", peers[i].Address)
	// }

	content += fmt.Sprintf("<h2>TxPool</h2>%d<br>", len(chain.Tx_pool))

	for i := 0; i < len(chain.Tx_pool); i++ {
		//content += fmt.Sprintf("Nonce %d, Id %x<br>", chain.Tx_pool[i].Nonce, chain.Tx_pool[i].Id[:])
		ctx := chain.Tx_pool[i]
		content += fmt.Sprintf("%d from %s to %s %x<br>", ctx.Amount, ctx.Sender, ctx.Receiver, ctx.Id)
	}

	content += fmt.Sprintf("<h2>Accounts</h2>number of accounts: %d<br><br>", len(chain.Accounts))

	for k, v := range chain.Accounts {
		content += fmt.Sprintf("%s %d<br>", k, v)
	}

	content += fmt.Sprintf("<br><h2>Blocks</h2><i>number of blocks %d</i><br>", len(chain.Blocks))

	for i := 0; i < len(chain.Blocks); i++ {
		t := chain.Blocks[i].Timestamp
		tsf := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())

		//summary
		content += fmt.Sprintf("<br><h3>Block %d</h3>timestamp %s<br>hash %x<br>prevhash %x\n", chain.Blocks[i].Height, tsf, chain.Blocks[i].Hash, chain.Blocks[i].Prev_Block_Hash)

		content += fmt.Sprintf("<h4>Number of Tx %d</h4>", len(chain.Blocks[i].Txs))
		for j := 0; j < len(chain.Blocks[i].Txs); j++ {
			ctx := chain.Blocks[i].Txs[j]
			content += fmt.Sprintf("%d from %s to %s %x<br>", ctx.Amount, ctx.Sender, ctx.Receiver, ctx.Id)
		}
	}

	return content
}

func Runweb(webport int) {
	//webserver to access node state through browser
	// HTTP
	log.Printf("start webserver %d", webport)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := LoadContent()
		//nlog.Print(p)
		fmt.Fprintf(w, "<h1>Polygon chain</h1><div>%s</div>", p)
	})

	nlog.Fatal(http.ListenAndServe(":"+strconv.Itoa(webport), nil))

}

//deal with the logic of each connection
//simple readwriter
// func (t *TCPServer) handleConnectionReadWriter(ntchan ntcl.Ntchan) {
// 	tr := 100 * time.Millisecond
// 	defer ntchan.Conn.Close()
// 	log.Println("handleConnection")

// 	for {

// 		log.Println("read with delim ", ntcl.DELIM)
// 		req, err := ntcl.NtwkRead(ntchan, ntcl.DELIM)

// 		if err != nil {
// 			log.Println(err)
// 		}

// 		if len(req) > 0 {
// 			log.Println("=> ", req, len(req))
// 			req = strings.Trim(req, string(ntcl.DELIM))
// 			resp := echohandler(req)

// 			log.Println("resp => ", resp)
// 			ntcl.NtwkWrite(ntchan, resp)

// 		} else {
// 			//empty read next read slower
// 			tr += 100 * time.Millisecond
// 		}

// 		time.Sleep(tr)
// 		//on empty reads increase time, but max at 800
// 		if tr > 800*time.Millisecond {
// 			tr = 800 * time.Millisecond
// 		}

// 	}
// }

// create a new Server
func NewServer(addr string) (*TCPServer, error) {
	return &TCPServer{
		addr:          addr,
		accepting:     false,
		ConnectedChan: make(chan net.Conn),
		//Peers: make([]ntcl.Peer)
	}, nil

}

// Close shuts down the TCP Server
func (t *TCPServer) Close() (err error) {
	return t.server.Close()
}

func setupLogfile() *log.Logger {
	//setup log file

	logFile, err := os.OpenFile(logfile_name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		nlog.Fatal(err)
	}

	//defer logfile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	logger := log.New(logFile, "node ", log.LstdFlags)
	logger.SetOutput(mw)

	//log.SetOutput(file)

	nlog = logger
	return logger

}

func run_node(config Configuration) {

	setupLogfile()

	nlog.Println("run node ", config.NodePort)

	// 	//TODO signatures of genesis
	chain.InitAccounts()

	// 	nlog.Println("PeerAddresses: ", config.PeerAddresses)

	success := chain.ReadChain()
	log.Println("read chain success ", success)
	nlog.Printf("block height %d", len(chain.Blocks))
	//chain.WriteGenBlock(chain.Blocks[0])

	// 	//create new genesis block (demo)
	createDemo := true //!success
	if createDemo {
		genBlock := chain.MakeGenesisBlock()
		chain.ApplyBlock(genBlock)
		chain.AppendBlock(genBlock)
	}

	// 	//if file exists read the chain

	// 	// create block every 10sec

	// 	delegation_enabled := false
	// 	if delegation_enabled {
	// 		go doEvery(blockTime, chain.MakeBlock)
	// 	}

	srv, err := NewServer(":" + strconv.Itoa(node_port))

	if err != nil {
		log.Println("error creating TCP server")
		return
	}

	// if err2 != nil {
	// 	log.Println("error starting TCP server ", err2)
	// 	return
	// }

	go srv.HandleConnect()

	srv.Run()
}

func LoadConfiguration(file string) Configuration {
	var config Configuration
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func main() {

	config := LoadConfiguration("nodeconf.json")

	go run_node(config)

	Runweb(config.WebPort)

}