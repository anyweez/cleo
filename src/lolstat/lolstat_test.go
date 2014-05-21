package main

import "testing"
import "fmt"

func test_compare(first []uint32, second []uint32) bool {
	if len(first) != len(second) {
		return false
	}
	
	for i := 0; i < len(first); i++ {
		if first[i] != second[i] {
			return false
		}
	}
	
	return true
} 

func TestOverlap1(t *testing.T) {
	a := []uint32{1, 2, 3, 4, 5}
	b := []uint32{2, 3, 4}
	
	overlap(&a, b)

	fmt.Println("Result:", a)
	if !test_compare(a, []uint32{2, 3, 4}) {
		t.Fail()
	}
}

func TestOverlap2(t *testing.T) {
	a := []uint32{1, 2, 3, 4, 5}
	b := []uint32{2, 4, 5}
	
	overlap(&a, b)

	if !test_compare(a, []uint32{2, 4, 5}) {
		t.Fail()
	}
}

func TestOverlap3(t *testing.T) {
	a := []uint32{1, 2, 63, 88, 110}
	b := []uint32{2, 4, 5}
	
	overlap(&a, b)

	if !test_compare(a, []uint32{2}) {
		t.Fail()
	}
}

func TestOverlap4(t *testing.T) {
	a := []uint32{1, 2, 63, 88, 110}
	b := []uint32{1}
	
	overlap(&a, b)

	if !test_compare(a, []uint32{1}) {
		t.Fail()
	}
}

func TestOverlap5(t *testing.T) {
	a := []uint32{1, 2, 63, 88, 110}
	b := []uint32{110}
	
	overlap(&a, b)

	if !test_compare(a, []uint32{110}) {
		t.Fail()
	}
}

func TestMerge1(t *testing.T) {
	a := []uint32{1, 5, 7}
	b := []uint32{2, 8, 10}
	
	if !test_compare( merge(a, b), []uint32{1, 2, 5, 7, 8, 10} ) {
		t.Fail()
	}
}

func TestMerge2(t *testing.T) {
	a := []uint32{1, 2, 3}
	b := []uint32{4, 5, 6}
	
	if !test_compare( merge(a, b), []uint32{1, 2, 3, 4, 5, 6} ) {
		t.Fail()
	}
}

func TestMerge3(t *testing.T) {
	a := []uint32{4, 5, 6}
	b := []uint32{1, 2, 3}
	
	if !test_compare( merge(a, b), []uint32{1, 2, 3, 4, 5, 6} ) {
		t.Fail()
	}
}
