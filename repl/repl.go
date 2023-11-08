package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/parser"
)

const PROMPT = ">> "
const MONKEY = `                  
                                   __                           
                                  /  |                          
_____  ____    ______   _______   $$ |   __   ______   __    __ 
/     \/    \  /      \ /       \ $$ |  /  | /      \ /  |  /  |
$$$$$$ $$$$  |/$$$$$$  |$$$$$$$  |$$ |_/$$/ /$$$$$$  |$$ |  $$ |
$$ | $$ | $$ |$$ |  $$ |$$ |  $$ |$$   $$<  $$    $$ |$$ |  $$ |
$$ | $$ | $$ |$$ \__$$ |$$ |  $$ |$$$$$$  \ $$$$$$$$/ $$ \__$$ |
$$ | $$ | $$ |$$    $$/ $$ |  $$ |$$ | $$  |$$       |$$    $$ |
$$/  $$/  $$/  $$$$$$/  $$/   $$/ $$/   $$/  $$$$$$$/  $$$$$$$ |
                                                      /  \__$$ |
                                                      $$    $$/
                                                       $$$$$$/ 

`

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, PROMPT)
		scanned:= scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.MakeNewParser(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParseError(out, p.Errors())
			continue
		}
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printParseError(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}