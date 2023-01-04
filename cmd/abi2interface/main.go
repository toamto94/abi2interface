package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
)

var name = flag.String("n", "iAbi2Interface", "name of the interface")
var inputFile = flag.String("f", "", "path to the input abi json file")
var outputFile = flag.String("o", "./interface.sol", "path to the output solidity interface file")
var solidityVersion = flag.String("v", "0.7.18", "solidity version")

type ABI []struct {
	Anonymous bool `json:"anonymous,omitempty"`
	Inputs    []struct {
		Indexed      bool   `json:"indexed"`
		InternalType string `json:"internalType"`
		Name         string `json:"name"`
		Type         string `json:"type"`
	} `json:"inputs"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Outputs []struct {
		InternalType string `json:"internalType"`
		Name         string `json:"name"`
		Type         string `json:"type"`
	} `json:"outputs,omitempty"`
	StateMutability string `json:"stateMutability,omitempty"`
}

func main() {
	flag.Parse()
	abi, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Not able to open file %v", *inputFile)
	}
	byteValue, _ := io.ReadAll(abi)

	abiStruct := ABI{}
	json.Unmarshal([]byte(byteValue), &abiStruct)

	o, err := os.OpenFile(*outputFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Not able to create file %v", *outputFile)
	}
	defer o.Close()

	license := []byte("// SPDX-License-Identifier: MIT\n\n")
	version := []byte("pragma solidity ^" + *solidityVersion + ";\n\n")

	if _, err := o.Write(license); err != nil {
		log.Fatalf("Could not write %v", *outputFile)
	}

	if _, err := o.Write(version); err != nil {
		log.Fatalf("Could not write %v", *outputFile)
	}

	if _, err := o.Write([]byte("interface " + *name + " {\n")); err != nil {
		log.Fatalf("Could not write %v", *outputFile)
	}

	for _, property := range abiStruct {
		if property.Type == "function" || property.Type == "event" {
			var s string
			s = "  " + property.Type + " "
			s = s + property.Name + "("
			for i, arg := range property.Inputs {
				if !arg.Indexed {
					if arg.Name == "" {
						s = s + arg.Type
					} else {
						s = s + arg.Type + " " + arg.Name
					}
				} else {
					s = s + arg.Type + " " + "indexed " + arg.Name
				}

				if i != len(property.Inputs)-1 {
					s = s + ", "
				}
			}
			s = s + ")"

			switch property.Type {
			case "function":
				s = s + " external "
				if property.StateMutability != "" {
					s = s + property.StateMutability
				}
				if len(property.Outputs) != 0 {
					s = s + " returns("
					for i, output := range property.Outputs {
						if output.Name == "" {
							s = s + output.Type
						} else {
							s = s + output.Type + " " + output.Name
						}
						if i != len(property.Outputs)-1 {
							s = s + ", "
						}
					}
					s = s + ");"
				} else {
					s = s + ";"
				}
			case "event":
				s = s + ";"
			}

			if _, err := o.Write([]byte(s + "\n")); err != nil {
				log.Fatalf("Could not write %v", *outputFile)
			}
		}

	}

	if _, err := o.Write([]byte("}")); err != nil {
		log.Fatalf("Could not write %v", *outputFile)
	}
	defer abi.Close()
}
