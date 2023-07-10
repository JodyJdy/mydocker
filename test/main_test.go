package test

import (
	"fmt"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	// ...

	fmt.Fprint(os.Stdout, "hello world")
}
