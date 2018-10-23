package gen_x64

import (
	"monkey/ast"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/parser"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

type testCase struct {
	input    string
	expected int
}

func TestGenerator(t *testing.T) {
	tests := []testCase{
		{
			input:    `return 1`,
			expected: 1,
		},
		{
			input:    `return 1 + 1`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		// parse
		program := parse(tt.input)

		// compile(to bytecode)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		// compile(x86 code generation)
		g := New(comp.Bytecode())
		err = g.Genx64()
		if err != nil {
			t.Errorf("code generation error: %s", err)
		}

		// write tmp file
		f, err := os.OpenFile("/tmp/monkeytmp.s", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			t.Errorf(err.Error())
		}
		f.Write(g.assembly.Bytes())
		f.Close()

		// change assembly to machine code and link.
		cmd := exec.Command("/usr/bin/gcc", "/tmp/monkeytmp.s", "-o", "/tmp/monkeytmp")
		_ = cmd.Run()
		cmd = exec.Command("/tmp/monkeytmp")
		err = cmd.Run()
		returncode := -1

		if err != nil {
			if s, ok := err.(*exec.ExitError).Sys().(syscall.WaitStatus); ok {
				returncode = s.ExitStatus()
			} else {
				t.Errorf("can't get return code")
			}
		} else {
			returncode = 1
		}

		if returncode != tt.expected {
			t.Errorf("return code is different got=%d, expected=%d", returncode, tt.expected)
		}
	}

	// delete gabages
	os.Remove("/tmp/monkeytmp.s")
	os.Remove("/tmp/mokeytmp")
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
