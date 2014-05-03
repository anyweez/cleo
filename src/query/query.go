package query

import "gamelog"

type FilterFunction func(*gamelog.GameRecord, *Query) bool

type Query struct {
	FilterText [2]string		// A bunch of strings that need to be present in order to keep a record
	LabelText [2]string
	Filters [1]FilterFunction
	Labelers [1]FilterFunction
}

func ReadFile(filename string) *Query {
	query := new(Query)
	query.FilterText[0] = "thresh"
	query.LabelText[0] = "thresh"
	
	// Check that everyone in Filters exists
	// query.Checks[0]
	query.Filters[0] = func(record *gamelog.GameRecord, query *Query) bool {
		// Check each team.
		for _, team := range record.Teams {
			// Check each player on each team.
			for _, player := range team.Players {
				// Check to see if the target champion exists there.
				for _, target := range query.FilterText {
					// If yes, return true.
					if target == *(player.Champion) {
						return true
					}
				}
			}
		}
		
		return false
	}
	
	query.Labelers[0] = func(record *gamelog.GameRecord, query *Query) bool {
		// Check each team.
		for _, team := range record.Teams {
			// Check each player on each team.
			for _, player := range team.Players {
				// Check to see if the target champion exists there.
				for _, target := range query.FilterText {
					// If yes, return true.
					if target == *(player.Champion) {
						return *team.Victory
					}
				}
			}
		}
		
		return false
	}
	
	return query
}

func prepare(raw_query string) {
	
}
