package query

//import "adapter_league" adapter
import "proto"
import "fmt"

type QueryManager struct {
		// Connection information.
}

func (q *QueryManager) Connect() {
	
}

func (q *QueryManager) Await() proto.GameQuery {
	query := proto.GameQuery{}
	query.Winners = append(query.Winners, proto.ChampionType_THRESH)
	
	return query
}

func (q *QueryManager) Respond(qr *proto.GameQueryResponse) {
	fmt.Println("Events examined:", qr.Total)
	fmt.Println(fmt.Sprintf("Matches: %d / %d", qr.Matching, qr.Available))
}
