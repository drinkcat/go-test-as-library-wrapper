package two_test

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	fmt.Printf("2 plus 2 is 3\n")
}

func TestMain(main *testing.M) {
	fmt.Printf("TestMain ran\n")

    main.Run()
}
