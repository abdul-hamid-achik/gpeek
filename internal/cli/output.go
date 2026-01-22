package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// outputJSON marshals data to JSON and prints it
func outputJSON(data interface{}) {
	var output []byte
	var err error

	if formatFlag == "compact" {
		output, err = json.Marshal(data)
	} else {
		output, err = json.MarshalIndent(data, "", "  ")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

// output handles outputting data based on format flag
func output(data interface{}, plainFormatter func(interface{}) string) {
	switch formatFlag {
	case "json", "compact":
		outputJSON(data)
	default:
		if plainFormatter != nil {
			fmt.Print(plainFormatter(data))
		} else {
			fmt.Printf("%+v\n", data)
		}
	}
}
