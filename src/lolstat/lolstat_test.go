package main

import "container/list"
import "testing"
import "libcleo"
import "fmt"
import "log"
import "strings"

func print_list(l *list.List) {
	values := make([]string, 0, 100)
	values = append(values, "list=")

	for iter := l.Front(); iter != nil; iter = iter.Next() {
		values = append(values, fmt.Sprintf("%s", (*iter).Value))
	}

	log.Println(strings.Join(values, ","))
}

func test_compare(first *list.List, second []libcleo.GameId) bool {
	if first.Len() != len(second) {
		return false
	}

	f_iter := first.Front()
	for i := 0; i < len(second); i++ {
		if (*f_iter).Value.(libcleo.GameId) != second[i] {
			return false
		}

		f_iter = f_iter.Next()
	}

	return true
}

func TestOverlap1(t *testing.T) {
	a := list.New()
	for i := 0; i < 10; i++ {
		a.PushBack(libcleo.GameId(i))
	}
	b := []libcleo.GameId{2, 3, 4}

	log.Println("Testing overlap.")
	print_list(a)
	overlap(a, b)
	print_list(a)

	log.Println("Overlap complete. Comparing results.")
	if !test_compare(a, []libcleo.GameId{2, 3, 4}) {
		t.Fail()
	}
}

func TestOverlap2(t *testing.T) {
	a := list.New()
	for i := 0; i < 10; i++ {
		a.PushBack(libcleo.GameId(i))
	}
	b := []libcleo.GameId{2, 4, 5}

	overlap(a, b)

	if !test_compare(a, []libcleo.GameId{2, 4, 5}) {
		t.Fail()
	}
}

func TestOverlap3(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(1))
	a.PushBack(libcleo.GameId(2))
	a.PushBack(libcleo.GameId(63))
	a.PushBack(libcleo.GameId(88))
	a.PushBack(libcleo.GameId(110))
	b := []libcleo.GameId{2, 4, 5}

	overlap(a, b)

	if !test_compare(a, []libcleo.GameId{2}) {
		t.Fail()
	}
}

func TestOverlap4(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(1))
	a.PushBack(libcleo.GameId(2))
	a.PushBack(libcleo.GameId(63))
	a.PushBack(libcleo.GameId(88))
	a.PushBack(libcleo.GameId(110))
	b := []libcleo.GameId{1}

	overlap(a, b)

	if !test_compare(a, []libcleo.GameId{1}) {
		t.Fail()
	}
}

func TestOverlap5(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(1))
	a.PushBack(libcleo.GameId(2))
	a.PushBack(libcleo.GameId(63))
	a.PushBack(libcleo.GameId(88))
	a.PushBack(libcleo.GameId(110))
	b := []libcleo.GameId{110}

	overlap(a, b)

	if !test_compare(a, []libcleo.GameId{110}) {
		t.Fail()
	}
}

func TestOverlap6(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(1371765903))
	a.PushBack(libcleo.GameId(1373700208))
	a.PushBack(libcleo.GameId(1381119672))
	a.PushBack(libcleo.GameId(1381634174))
	a.PushBack(libcleo.GameId(1386748546))
	a.PushBack(libcleo.GameId(1386956597))
	a.PushBack(libcleo.GameId(1389613363))
	a.PushBack(libcleo.GameId(1389891620))
	a.PushBack(libcleo.GameId(1390726214))
	a.PushBack(libcleo.GameId(1391074822))
	a.PushBack(libcleo.GameId(1392060257))
	a.PushBack(libcleo.GameId(1392233260))
	a.PushBack(libcleo.GameId(1392837812))
	b := []libcleo.GameId{1248544229, 1286363507, 1365516208, 1368006647, 1371765903, 1371835613, 1372862292, 1372951151, 1373097982, 1373147336, 1373700208, 1373731529, 1373744617, 1373752600, 1373757764, 1378584683, 1380075900, 1380156021, 1381119672, 1381169485, 1381224811, 1381246301, 1381299765, 1381569988, 1381629953, 1381634174, 1381666055, 1381722684, 1383371256, 1383555399, 1383662939, 1383693865, 1383743891, 1383770286, 1383791214, 1383796189, 1383922458, 1383958971, 1384021108, 1386748546, 1386805916, 1386956597, 1387053594, 1387694953, 1387965849, 1389534165, 1389613363, 1389880058, 1389882116, 1389882652, 1389889872, 1389891620, 1390057392, 1390390779, 1390403495, 1390617312, 1390726214, 1390728558, 1390760412, 1391050694, 1391051650, 1391062065, 1391070341, 1391074822, 1391085759, 1391093769, 1391109966, 1391148381, 1391159649, 1391165680, 1391171156, 1391174017, 1391197284, 1391198309, 1391265671, 1391468566, 1392055317, 1392060257, 1392080014, 1392129263, 1392164182, 1392182298, 1392197279, 1392233260, 1392252045, 1392268716, 1392279461, 1392280810, 1392687017, 1392708282, 1392760773, 1392799368, 1392837812, 1393063887, 1393074710, 1393095734, 1393140131, 1393150462, 1393178637, 1393252751}

	overlap(a, b)

	if !test_compare(a, []libcleo.GameId{1371765903, 1373700208, 1381119672, 1381634174, 1386748546, 1386956597, 1389613363, 1389891620, 1390726214, 1391074822, 1392060257, 1392233260, 1392837812}) {
		t.Fail()
	}
}

func TestMerge1(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(1))
	a.PushBack(libcleo.GameId(5))
	a.PushBack(libcleo.GameId(7))

	b := list.New()
	b.PushBack(libcleo.GameId(2))
	b.PushBack(libcleo.GameId(8))
	b.PushBack(libcleo.GameId(10))

	if !test_compare(merge(a, b), []libcleo.GameId{1, 2, 5, 7, 8, 10}) {
		t.Fail()
	}
}

func TestMerge2(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(1))
	a.PushBack(libcleo.GameId(2))
	a.PushBack(libcleo.GameId(3))

	b := list.New()
	b.PushBack(libcleo.GameId(4))
	b.PushBack(libcleo.GameId(5))
	b.PushBack(libcleo.GameId(6))

	if !test_compare(merge(a, b), []libcleo.GameId{1, 2, 3, 4, 5, 6}) {
		t.Fail()
	}
}

func TestMerge3(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(4))
	a.PushBack(libcleo.GameId(5))
	a.PushBack(libcleo.GameId(6))

	b := list.New()
	b.PushBack(libcleo.GameId(1))
	b.PushBack(libcleo.GameId(2))
	b.PushBack(libcleo.GameId(3))

	if !test_compare(merge(a, b), []libcleo.GameId{1, 2, 3, 4, 5, 6}) {
		t.Fail()
	}
}

func TestMerge4(t *testing.T) {
	a := list.New()
	a.PushBack(libcleo.GameId(3))
	a.PushBack(libcleo.GameId(4))
	a.PushBack(libcleo.GameId(5))

	b := list.New()
	b.PushBack(libcleo.GameId(3))
	b.PushBack(libcleo.GameId(5))
	b.PushBack(libcleo.GameId(6))

	if !test_compare(merge(a, b), []libcleo.GameId{3, 4, 5, 6}) {
		t.Fail()
	}
}
