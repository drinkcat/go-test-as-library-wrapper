package main

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"positive numbers", 1, 2, 3},
		{"negative numbers", -1, -2, -3},
		{"mixed numbers", -1, 2, 1},
		{"zero and positive", 0, 5, 5},
		{"zero and negative", -5, 0, -5},
		{"zero and zero", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Add(tt.a, tt.b); got != tt.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMain(main *testing.M) {
	fmt.Printf("TestMain ran\n")

    main.Run()
}
