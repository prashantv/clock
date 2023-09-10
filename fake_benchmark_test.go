package clock

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkFake1Timer(b *testing.B) {
	f := NewFake()
	t := f.Clock.Ticker(time.Millisecond)
	for i := 0; i < b.N; i++ {
		f.Add(time.Millisecond)
		<-t.C
	}
}

func BenchmarkSliceGrow(b *testing.B) {
	s := make([]int, 0, 1)

	for i := 0; i < b.N; i++ {
		s = append(s, 1)
		s = s[1:]
	}
}

func TestFoo(t *testing.T) {
	var s []int
	s = make([]int, 0, 2)

	add := func(i int) {
		s = append(s, i)
		fmt.Println("after", i, "cap", cap(s))
	}

	fmt.Println("initial capacity", cap(s))
	add(1)
	s = s[1:]
	add(2)
	s = s[1:]
	add(3)
	s = s[1:]
	add(4)
	s = s[1:]
	add(5)
}
