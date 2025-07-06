package main

type Message struct {
	Header
	Packets []Packet
}

type Header struct {
	seqNo      uint32
	msgID      uint64
	numPackets uint32
}

type Packet struct {
	seqNo  uint32
	length uint16
	data   []byte
}

type Server struct {
	listener *net.UDPListener
}

func main() {
	ctx := context.TODO()

	server := &Server{}

connLoop:
	for {
		select {
		case <-ctx.Done():
			break connLoop
		default:
			conn := listener.Accept()
			go handleConn(ctx, conn)
		}
	}
}

func handleConn(ctx context.Context, conn *net.UDPConn) {
	// TODO: recover

	defer conn.Close()

	var b [unsafe.Sizeof(Header)]byte

	_, err := conn.Read(b[:])
	if err != nil {
		panic(err)
	}

	header := *(*Header)(unsafe.Pointer(&b[0]))

	if header.seqNo != 0 ||
		header.msgID == 0 ||
		header.numPackets == 0 {
		return
	}

	for range header.numPackets {
		// TODO: read max package size for each packet to buffer
	}

	// TODO: decode buffer into packets
	// TODO: decode packets into message
	// TODO: decode message into request
}
