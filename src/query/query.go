package query

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"log"
	"net"
	"proto"
	"switchboard"
	"time"
)

type QueryRequest struct {
	Query	interface{}
	Conn	*net.Conn

	TimeReceived	int64
}

type GameQueryRequest struct {
	Id    string
	Query *proto.GameQuery

	Identity   string
	Connection *net.Conn
}

type GameQueryResponse struct {
	Response *proto.QueryResponse
	Request  *GameQueryRequest
}

type QueryManager struct {
	// Connection information.
	ActiveCount uint32

	//	Listener *net.TCPListener
	Switchboard switchboard.SwitchboardServer
}

func GetQueryId(qry proto.GameQuery) string {
	return fmt.Sprintf("Q%d.%d", qry.QueryProcess, qry.QueryId)
}

func (q *QueryManager) Connect(port int) {
	q.ActiveCount = 0
	cerr := error(nil)

	q.Switchboard, cerr = switchboard.NewServer("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: port})

	if cerr != nil {
		log.Fatal("Couldn't open port for listening.")
	}

	log.Println("Query server listening on port" + string(port))
}

func (q *QueryManager) Listen(query_type gproto.Message) QueryRequest {
	query := QueryRequest{}
	log.Println("Awaiting request")

	conn, _ := q.Switchboard.GetStream()
	query.Conn = conn
	rw := bufio.NewReadWriter(bufio.NewReader(*conn), bufio.NewWriter(*conn))

	msg, _ := rw.ReadString('|')
	merr := gproto.Unmarshal([]byte(msg[:len(msg)-1]), query_type)

	if merr != nil {
		log.Fatal("Error unmarshaling query.")
	}

	query.Query = query_type
	query.TimeReceived = time.Now().Unix()

	return query
}

// TODO: get rid of await in favor of a more generic Listen
func (q *QueryManager) Await() GameQueryRequest {
	gqr := GameQueryRequest{}

	log.Println("Awaiting request.")
	conn, _ := q.Switchboard.GetStream()

	gqr.Connection = conn
	rw := bufio.NewReadWriter(bufio.NewReader(*conn), bufio.NewWriter(*conn))

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

	// Print out the included champions for debugging purposes.
	champ_str := "allies="
	for _, ch := range qry.Winners {
		champ_str += ch.String() + ","
	}
	champ_str += " enemies="
	for _, ch := range qry.Losers {
		champ_str += ch.String() + ","
	}
	log.Println(fmt.Sprintf("%s: requires champion %s", gqr.Id, champ_str))

	// Increment the query counter.
	q.ActiveCount += 1

	return gqr
}

func (q *QueryManager) Respond(qr *GameQueryResponse) {
	defer (*qr.Request.Connection).Close()

	data, _ := gproto.Marshal(qr.Response)

	// Send the data back to the responder and decrement the # of active queries.
	rw := bufio.NewReadWriter(bufio.NewReader(*qr.Request.Connection), bufio.NewWriter(*qr.Request.Connection))
	rw.WriteString(string(data) + "|")
	rw.Flush()
	log.Println(fmt.Sprintf("%s: sent response", qr.Request.Id))

	q.ActiveCount -= 1
}
