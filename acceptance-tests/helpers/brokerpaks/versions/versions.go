package main

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/brokerpaks"
)

func main() {
	for _, v := range brokerpaks.Versions() {
		fmt.Println(v)
	}
}
