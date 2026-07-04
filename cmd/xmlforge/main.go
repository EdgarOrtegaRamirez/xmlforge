// XmlForge CLI - XML Processing Toolkit
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/convert"
	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/diff"
	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/format"
	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/parser"
	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/stats"
	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/validator"
	"github.com/EdgarOrtegaRamirez/xmlforge/pkg/xpath"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		cmdParse()
	case "format":
		cmdFormat()
	case "compress":
		cmdCompress()
	case "stats":
		cmdStats()
	case "xpath":
		cmdXPath()
	case "diff":
		cmdDiff()
	case "convert":
		cmdConvert()
	case "validate":
		cmdValidate()
	case "version":
		fmt.Printf("xmlforge v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`XmlForge - XML Processing Toolkit v` + version + `

Usage:
  xmlforge <command> [options] <file>

Commands:
  parse       Parse and display XML structure
  format      Pretty-print XML
  compress    Remove whitespace from XML
  stats       Show XML document statistics
  xpath       Query XML using XPath expressions
  diff        Compare two XML files
  convert     Convert XML to other formats
  validate    Validate XML against rules
  version     Show version
  help        Show this help message

Examples:
  xmlforge parse input.xml
  xmlforge format input.xml
  xmlforge stats input.xml
  xmlforge xpath "root/child" input.xml
  xmlforge diff file1.xml file2.xml
  xmlforge convert --to json input.xml
  xmlforge validate --max-depth 5 input.xml`)
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(data), nil
}

func cmdParse() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge parse <file>\n")
		os.Exit(1)
	}

	xml, err := readFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	doc, err := parser.ParseString(xml)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Root: <%s>\n", doc.Root.Name)
	if len(doc.Root.Attributes) > 0 {
		fmt.Println("Attributes:")
		for _, attr := range doc.Root.Attributes {
			fmt.Printf("  %s=%q\n", attr.Name, attr.Value)
		}
	}
	printNode(doc.Root, "  ")
}

func printNode(node *parser.Node, indent string) {
	for _, child := range node.Children {
		switch child.Type {
		case parser.NodeElement:
			fmt.Printf("%s<%s>", indent, child.Name)
			if len(child.Attributes) > 0 {
				for _, attr := range child.Attributes {
					fmt.Printf(" %s=%q", attr.Name, attr.Value)
				}
			}
			if text := child.GetText(); text != "" && len(child.Children) == 0 {
				fmt.Printf("%s</%s>\n", text, child.Name)
			} else {
				fmt.Println()
				if text := child.GetText(); text != "" {
					fmt.Printf("%s  %s\n", indent, text)
				}
				printNode(child, indent+"  ")
				fmt.Printf("%s</%s>\n", indent, child.Name)
			}
		case parser.NodeText:
			fmt.Printf("%sText: %s\n", indent, strings.TrimSpace(child.Value))
		case parser.NodeComment:
			fmt.Printf("%s<!-- %s -->\n", indent, child.Value)
		}
	}
}

func cmdFormat() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge format <file>\n")
		os.Exit(1)
	}

	xml, err := readFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result, err := format.Format(xml, format.DefaultOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func cmdCompress() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge compress <file>\n")
		os.Exit(1)
	}

	xml, err := readFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := format.Compress(xml)
	fmt.Println(result)
}

func cmdStats() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge stats <file>\n")
		os.Exit(1)
	}

	xml, err := readFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	doc, err := parser.ParseString(xml)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	s := stats.Analyze(doc)
	fmt.Print(s.Format())
}

func cmdXPath() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge xpath <expression> <file>\n")
		os.Exit(1)
	}

	expr := os.Args[2]
	xml, err := readFile(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	doc, err := parser.ParseString(xml)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	results, err := xpath.Execute(doc, expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "XPath error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(xpath.String(results))
}

func cmdDiff() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge diff <file1> <file2>\n")
		os.Exit(1)
	}

	xml1, err := readFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", os.Args[2], err)
		os.Exit(1)
	}

	xml2, err := readFile(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", os.Args[3], err)
		os.Exit(1)
	}

	result, err := diff.CompareStrings(xml1, xml2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Diff error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result.Format())
}

func cmdConvert() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge convert --to <format> <file>\n")
		os.Exit(1)
	}

	formatFlag := os.Args[2]
	toFormat := ""
	if formatFlag == "--to" && len(os.Args) > 3 {
		toFormat = os.Args[3]
	}

	filePath := ""
	if toFormat != "" {
		filePath = os.Args[4]
	} else {
		filePath = os.Args[2]
	}

	if filePath == "" {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge convert --to <format> <file>\n")
		os.Exit(1)
	}

	xml, err := readFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	doc, err := parser.ParseString(xml)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	switch toFormat {
	case "json", "":
		result, err := convert.ToJSON(doc, convert.DefaultJSONOptions())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Convert error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case "csv":
		result, err := convert.ToCSV(doc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Convert error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(result)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported format: %s (supported: json, csv)\n", toFormat)
		os.Exit(1)
	}
}

func cmdValidate() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: xmlforge validate [options] <file>\n")
		os.Exit(1)
	}

	rules := validator.DefaultRules()

	// Parse options
	filePath := os.Args[len(os.Args)-1]
	for i := 2; i < len(os.Args)-1; i++ {
		switch os.Args[i] {
		case "--max-depth":
			if i+1 < len(os.Args)-1 {
				fmt.Sscanf(os.Args[i+1], "%d", &rules.MaxDepth)
				i++
			}
		case "--max-attrs":
			if i+1 < len(os.Args)-1 {
				fmt.Sscanf(os.Args[i+1], "%d", &rules.MaxAttributes)
				i++
			}
		case "--no-empty":
			rules.NoEmptyElements = true
		case "--check-names":
			rules.ValidateNames = true
		}
	}

	xml, err := readFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result, err := validator.ValidateString(xml, rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result.Format())

	if !result.Valid {
		os.Exit(1)
	}
}
