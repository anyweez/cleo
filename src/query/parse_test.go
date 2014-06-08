package query

import "testing"
import "io/ioutil"

func TestParse1(t *testing.T) {
	fp := io.Open("queries/thresh.lkg")
	query_text := string(ioutil.ReadAll(fp))

	parse_tree := Parse(query_text)
}
