package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Specify a conversion type, e.g., go2proto or sql2gorm")
		return
	}

	switch os.Args[1] {
	case "go2proto":
		go2protoCmd()
	// other conversion commands can be added here...
	default:
		fmt.Printf("Unsupported conversion type: %s\n", os.Args[1])
	}
}

func go2protoCmd() {
	cmd := flag.NewFlagSet("go2proto", flag.ExitOnError)
	goFile := cmd.String("src", "", "Source Go file to be converted to Proto format")
	outputFile := cmd.String("output", "", "Output file to write the result. Default is stdout.")
	cmd.Parse(os.Args[2:])

	if *goFile == "" {
		fmt.Println("Please provide a Go file using -src flag")
		return
	}

	result := convertGoFileToProto(*goFile)
	if strings.HasSuffix(*outputFile, ".proto") {
		result = "syntax = \"proto3\";\n\n" + result
	}

	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(result), 0644)
		if err != nil {
			fmt.Println("Error writing to output file:", err)
			return
		}
		fmt.Printf("Output written to file: %s\n", *outputFile)
	} else {
		fmt.Println(result)
	}
}
