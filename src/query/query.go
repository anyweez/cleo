package query

import (
	"fmt"
	"proto"
)

// TODO: replace this with something real.
type Connection struct {
	
}

type GameQueryRequest struct {
	Id 			uint32
	Conn		*Connection	
	Query 		*proto.GameQuery
}

type GameQueryResponse struct {
	Id			uint32
	Conn		*Connection
	Response	*proto.QueryResponse
}

type QueryManager struct {
	// Connection information.
	NextQueryId		uint32
	ActiveCount		uint32
}

func (q *QueryManager) Connect() {
	q.NextQueryId = 1
	q.ActiveCount = 0
}

func (q *QueryManager) Await() GameQueryRequest {
	gqr := GameQueryRequest{Id: q.NextQueryId}
	
	// Find out how often Thresh is on the winning team. This is a placeholder
	// query and should be replaced with a network listener.
	gqr.Query = &proto.GameQuery{ Winners: []proto.ChampionType{proto.ChampionType_THRESH}, Losers: []proto.ChampionType{} }
	
	// Increment the query counter.
	q.NextQueryId += 1
	q.ActiveCount += 1
	
	return gqr
}

func (q *QueryManager) Respond(qr *GameQueryResponse) {
	fmt.Println("Events examined:", *qr.Response.Total)
	fmt.Println(fmt.Sprintf("Matches: %d / %d [%.2f]", *qr.Response.Matching, *qr.Response.Available, float32(*qr.Response.Matching) / float32(*qr.Response.Available)))
	
	q.ActiveCount -= 1
}
