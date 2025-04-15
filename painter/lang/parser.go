package lang

import (
	"bufio"
	"io"
	"log"
	"strings"

	"github.com/gothicenemy/software-architecture-3/painter"
)

type Parser struct{}

func (p *Parser) Parse(r io.Reader) ([]painter.Operation, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	var res []painter.Operation
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		commandLine := scanner.Text()
		lineForParsing := commandLine

		// !! Видалення коментаря перед обробкою !!
		if commentIndex := strings.Index(lineForParsing, "#"); commentIndex != -1 {
			lineForParsing = lineForParsing[:commentIndex]
		}
		// -------------------------------------------------

		trimmedLine := strings.TrimSpace(lineForParsing)

		if trimmedLine == "" || strings.HasPrefix(strings.TrimSpace(commandLine), "#") {
			continue
		}

		op := parseLine(trimmedLine)
		if op != nil {
			res = append(res, op)
		} else {
			log.Printf("Warning: Skipping invalid command line %d: %s", lineNum, commandLine)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
		return nil, err
	}

	log.Printf("Parsing finished. Found %d valid operations.", len(res))
	return res, nil
}

func parseLine(line string) painter.Operation {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}

	command := strings.ToLower(fields[0])
	args := fields[1:]

	switch command {
	case "white":
		if len(args) != 0 {
			log.Println("Error parsing 'white': expects 0 arguments")
			return nil
		}
		return painter.WhiteOperation{}
	case "green":
		if len(args) != 0 {
			log.Println("Error parsing 'green': expects 0 arguments")
			return nil
		}
		return painter.GreenOperation{}
	case "update":
		if len(args) != 0 {
			log.Println("Error parsing 'update': expects 0 arguments")
			return nil
		}
		return painter.UpdateOperation{}
	case "bgrect":
		coords, ok := painter.ParseCoords(args, 4)
		if !ok {
			return nil
		}
		if coords[0] >= coords[2] || coords[1] >= coords[3] {
			log.Printf("Warning: Invalid bgrect coordinates %.2f,%.2f -> %.2f,%.2f (x1>=x2 or y1>=y2)", coords[0], coords[1], coords[2], coords[3])
		}
		return painter.BgRectOperation{X1: coords[0], Y1: coords[1], X2: coords[2], Y2: coords[3]}
	case "figure":
		coords, ok := painter.ParseCoords(args, 2)
		if !ok {
			return nil
		}
		return painter.FigureOperation{X: coords[0], Y: coords[1]}
	case "move":
		coords, ok := painter.ParseCoords(args, 2)
		if !ok {
			return nil
		}
		return painter.MoveOperation{X: coords[0], Y: coords[1]}
	case "reset":
		if len(args) != 0 {
			log.Println("Error parsing 'reset': expects 0 arguments")
			return nil
		}
		return painter.ResetOperation{}
	default:
		log.Printf("Error: Unknown command '%s'", command)
		return nil
	}
}
