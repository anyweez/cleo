package query

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"log"
	"net"
	"proto"
)

type GameQueryRequest struct {
	Id    string
	Query *proto.GameQuery

	Identity   string
	Connection *net.TCPConn
}

type GameQueryResponse struct {
	Response *proto.QueryResponse
	Request  *GameQueryRequest
}

type QueryManager struct {
	// Connection information.
	ActiveCount uint32

	Listener *net.TCPListener
}

func GetQueryId(qry proto.GameQuery) string {
	return fmt.Sprintf("Q%d.%d", qry.QueryProcess, qry.QueryId)
}

func (q *QueryManager) Connect() {
	q.ActiveCount = 0

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: 14002})
	log.Println("Listening on port 14002")

	q.Listener = listener

	if err != nil {
		log.Fatal("Couldn't open port 14002 for listening.")
	}

	log.Println("Query server listening on port 14002")
}

func (q *QueryManager) Await() GameQueryRequest {
	gqr := GameQueryRequest{}

	log.Println("Awaiting request.")
	conn, _ := (*q.Listener).AcceptTCP()

	gqr.Connection = conn
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	msg, _ := rw.ReadString('|')
	log.Println(fmt.Sprintf("Request received (%d bytes)", len(msg)))

	qry := proto.GameQuery{}
	merr := gproto.Unmarshal([]byte(msg[:len(msg)-1]), &qry)

	if merr != nil {
		log.Fatal("Error unmarshaling query from frontend.")
	}

	// Form a query ID string that can be used for logging.
	gqr.Id = GetQueryId(qry)
	gqr.Query = &qry

	for _, q := range qry.Winners {
		log.Println(fmt.Sprintf("%s: requires champion %s", gqr.Id, q))
	}

	// Increment the query counter.
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
	log.Println(fmt.Sprintf("%s: sent response", qr.Request.Id))

	q.ActiveCount -= 1
}
