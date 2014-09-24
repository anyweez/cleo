package query

import (
	"bufio"
	gproto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"log"
	"net"
	//	"proto"
	"switchboard"
	"time"
)

type QueryRequest struct {
	Query interface{}
	Conn  *net.Conn

	TimeReceived int64
}

/*
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
*/

type QueryManager struct {
	// Connection information.
	ActiveCount uint32

	//	Listener *net.TCPListener
	Switchboard switchboard.SwitchboardServer
}

/*
func GetQueryId(qry proto.GameQuery) string {
	return fmt.Sprintf("Q%d.%d", qry.QueryProcess, qry.QueryId)
}
*/
func (q *QueryManager) Connect(port int) {
	q.ActiveCount = 0
	cerr := error(nil)

	q.Switchboard, cerr = switchboard.NewServer("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: port})

	if cerr != nil {
		log.Fatal("Couldn't open port for listening.")
	}

	log.Println(fmt.Sprintf("Query server listening on port %d", port))
}

func (q *QueryManager) Listen(query_type gproto.Message) QueryRequest {
	query := QueryRequest{}

	conn, _ := q.Switchboard.GetStream()
	query.Conn = conn
	rw := bufio.NewReadWriter(bufio.NewReader(*conn), bufio.NewWriter(*conn))

	msg, _ := rw.ReadString('|')
	merr := gproto.Unmarshal([]byte(msg[:len(msg)-1]), query_type)

	if merr != nil {
		log.Fatal("Error unmarshaling query; are you sure you sending a properly formatted proto?")
	}

	query.Query = query_type
	query.TimeReceived = time.Now().Unix()

	return query
}

func (q *QueryManager) Reply(request *QueryRequest, response gproto.Message) {
	defer (*request.Conn).Close()

	data, _ := gproto.Marshal(response)

	// Send the data back to the responder and decrement the # of active queries.
	rw := bufio.NewReadWriter(bufio.NewReader(*request.Conn), bufio.NewWriter(*request.Conn))
	rw.WriteString(string(data) + "|")
	rw.Flush()
	log.Println("response sent")
}
