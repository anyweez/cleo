package query

/*
import "errors"

const ParseOperation (
	PARSE_SELECT = iota
)

type ParseTree struct {
	Operation ParseOperation
}

func Parse(text string) ParseTree {
	tokens := tokenize(text)

	parse_tree := ParseTree{}
	err := pStart(tokens, &parse_tree)

	return parse_tree
}

func pStart() (tokens []string, parse_tree *ParseTree) *ParseTree, errors.error {
	// The only valid pStart token is "select"
	if tokens[0] == "select" {
		parse_tree.Operation = PARSE_SELECT

		pt, err = pSelectors(tokens[1:])
		if pt == nil {
			return _, err
		}
	} else {
		return nil, errors.New("Invalid starting token.")
	}
}

func pSelectors(tokens []string, parse_tree *ParseTree) *ParseTree, errors.error {
	if tokens[1] == "contains" || tokens[1] == "has" || {

	} else {

	}

}

func pEnd(tokens []string, parse_tree *ParseTree) *ParseTree, errors.error {
	if tokens[0] == ";" {
		return parse_tree, nil
	} else {
		return nil, nil
	}
}


func tokenize(full string) []string {
	return nil
}
*/
