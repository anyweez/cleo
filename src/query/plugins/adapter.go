package lolstat

type QueryAdapter interface {
	Produce(filter_name string, params ...string)
}
