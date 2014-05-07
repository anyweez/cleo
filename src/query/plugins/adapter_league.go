package query

import "gamelog"
import "query"

type LeagueFilter func(team *gamelog.Team, filter_text string) bool
type LeagueSelector func(record *gamelog.GameRecord) *gamelog.Team

// Selectors are used to select a component from the GameRecord that filters
// should apply to. For example, selectors are used to identify what team
// "winner" refers to.
func Select() map[string]func {
	selectors = make( map[string]func )
	
	selectors["winner"] = func(record *gamelog.GameRecord) {
		for _, team := range record.Teams {
			if team.Victory {
				return team
			}
		}
		return nil
	}

	selectors["loser"] = func(record *gamelog.GameRecord) {
		for _, team := range record.Teams {
			if !team.Victory {
				return team
			}
		}
		return nil
	}
		
	return selectors
}

// Filters on selected components determine whether they should be included
// in filtered and labeled sets. Produce generates functions that can be
// used to filter selectors.
func Produce(filter_name string, params ...string) LeagueFilter {
	switch {
		// Return a function that can be used to identify whether a particular
		// team won.
		case filter_name == "champion":
			return func(team *gamelog.Team, cid uint32) bool {
				for _, player := team.Players {
					if player.champion_id == cid {
						return true
					}
				}
				return false
			}
			// return a function that accepts a single string param and returns a bool		
	}
}

func ParseText(filename string) {
	fp := io.Open(filename)
	
	new_query := query.Query()
	selectors = Select()
	
	line, err := fp.ReadString("\n")
	for err == nil {
		// Parse out the target
		segments := strings.Split(line, ":")
		selector := selectors[segments[0]]
		
		// Then get the filter
		split_loc := strings.Index(segments[1], "(")
		filter = Produce(segments[1][:split_loc], segments[1][split_loc+1:-1])
		
		new_query.QueryPairs = append(new_query.Statements, QueryStatement{selector, filter})
		
		line, err = fp.ReadString("\n")
	}
}
