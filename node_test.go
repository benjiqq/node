package main

//TODO!
// func TestBasicPing(t *testing.T) {

// 	msg_out_chan := make(chan protocol.Message)

// 	go HandlePing(msg_out_chan)
// 	msg := <-msg_out_chan
// 	msgString := protocol.MsgString(msg)

// 	if !(msgString == "REP#PONG#EMPTY|") {
// 		t.Error("ping failed ", msg)
// 	}
// }

// func TestBasicPingMsg(t *testing.T) {

// 	setupLogfile()

// 	msg_in := make(chan protocol.Message)
// 	msg_out := make(chan protocol.Message)

// 	emptydata := ""
// 	//req_msg := protocol.EncodeMsgString(protocol.REQ, protocol.CMD_PING, emptydata)
// 	req_msg := protocol.EncodeMsg(protocol.REQ, protocol.CMD_PING, emptydata)
// 	go func() {
// 		msg_in <- req_msg
// 	}()
// 	go HandleMsg(msg_in, msg_out)
// 	msg := <-msg_out
// 	msgString := protocol.MsgString(msg)
// 	if !(msgString == "REP#PONG#EMPTY|") {
// 		t.Error("ping failed ", msg)
// 	}
// }
