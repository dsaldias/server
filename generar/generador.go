package main

import (
	"fmt"
	"os"
)

func Init() {
	module := getModuleName()

	content := fmt.Sprintf(`package main

import (
	"net/http"
	"%s/graph"
)

func main() {
	_ = http.ListenAndServe(":8080", nil)
}
`, module)

	err := os.WriteFile("serverx.go", []byte(content), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("serverx.go creado ✨")
}
