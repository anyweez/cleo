package query

//import "adapter_league" adapter
import "proto"

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
}

func (q *QueryManager) Connect() {
	
	q.NextQueryId = 1
}

func (q *QueryManager) Await() GameQueryRequest {
	gqr := GameQueryRequest{}
	gqr.Id = q.NextQueryId
	
	// Find out how often Thresh is on the winning team.
	gqr.Query.Winners = append(gqr.Query.Winners, proto.ChampionType_THRESH)
	
	// Increment the query counter.
	q.NextQueryId += 1
	
	return gqr
}

/*
func (q *QueryManager) Respond(qr *lolstat.GameQueryResponse) {
	fmt.Println("Events examined:", qr.Total)
	fmt.Println(fmt.Sprintf("Matches: %d / %d", qr.Matching, qr.Available))
}
* */
