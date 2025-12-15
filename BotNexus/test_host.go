package main

import (
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v3/host"
)

func _main() {
	info, err := host.Info()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	b, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(b))
}
