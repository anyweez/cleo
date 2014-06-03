package query

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"log"
	"net"
	"proto"
//	zmq "github.com/pebbe/zmq4"
)

type GameQueryRequest struct {
	Id 			uint32
	Query 		*proto.GameQuery
	
	Identity	string
	Connection	*net.TCPConn
}

type GameQueryResponse struct {
	Id			uint32
	
	Response	*proto.QueryResponse	
	Request		*GameQueryRequest
}

type QueryManager struct {
	// Connection information.
	NextQueryId		uint32
	ActiveCount		uint32

	Listener		*net.TCPListener
}

func (q *QueryManager) Connect() {
	q.NextQueryId = 1
	q.ActiveCount = 0	
	
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP:net.IPv4zero, Port:14002})
	log.Println("Listening on port 14002")

	q.Listener = listener
	
	if err != nil {
		log.Fatal("Couldn't open port 14002 for listening.")
	}
	
	log.Println("Query server listening on port 14002")
}

func (q *QueryManager) Await() GameQueryRequest {
	gqr := GameQueryRequest{Id: q.NextQueryId}

	log.Println("Awaiting request.")
	conn, _ := (*q.Listener).AcceptTCP()

	gqr.Connection = conn
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	msg, _ := rw.ReadString('|')
	log.Println(fmt.Sprintf("Request received (%d bytes)", len(msg)))
	
	query := proto.GameQuery{}
	merr := gproto.Unmarshal( []byte(msg[:len(msg)-1]), &query )
	
	if merr != nil {
		log.Fatal("Error unmarshaling query from frontend.")
	}
	
	gqr.Query = &query
	
	for _, qry := range query.Winners {
		log.Println("Requires champion", qry)
	}
	
	// Increment the query counter.
	q.NextQueryId += 1
	q.ActiveCount += 1
	
	return gqr
}

func (q *QueryManager) Respond(qr *GameQueryResponse) {
//	defer (*qr.Request.Connection).Close()
	
	data, _ := gproto.Marshal(qr.Response)
	
	// Send the data back to the responder and decrement the # of active queries.
	rw := bufio.NewReadWriter(bufio.NewReader(qr.Request.Connection), bufio.NewWriter(qr.Request.Connection))
	rw.WriteString(string(data) + "|")
	rw.Flush()
	log.Println("Sent response.")
	
	q.ActiveCount -= 1
}
