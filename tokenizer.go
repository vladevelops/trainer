package main

import (
	"fmt"
	"os"
	"strings"
)

type Tokenizer struct {
	tokens          []string
	parser_position int
}

const TOKEN_EOF = "EOF"

func (t *Tokenizer) PullToken() (token string) {

	if t.parser_position > len(t.tokens)-1 {
		return TOKEN_EOF
	}

	current_token := t.tokens[t.parser_position]
	t.parser_position++
	return current_token
}

func (t *Tokenizer) CheckCurentToken() (token string) {
	return t.tokens[t.parser_position]
}

func (t *Tokenizer) tokenize_entire_file(path string) (err error) {

	f, read_file_err := os.ReadFile(path)

	if read_file_err != nil {

		return read_file_err
	}
	t.chars_to_tokens(strings.Split(string(f), ""))
	return nil
}

func (t *Tokenizer) print_tokens() {
	fmt.Printf("Parsed tokens: %#v \n", t.tokens)
}

type KEYWORD string

const (
	PHASES KEYWORD = "PHASES"
)

func (t *Tokenizer) chars_to_tokens(chars []string) {
	accomulator := ""
	char_index := 0
	for {
		if char_index == len(chars)-1 {
			break
		}
		char := chars[char_index]

		// t.print_tokens()
		// fmt.Printf("char: %#v char_index: %v \n", char, char_index)
		switch accomulator {
		case string(PHASES):
			t.tokens = append(t.tokens, string(PHASES))
			accomulator = ""
		case "#":
			for {
				char = chars[char_index]
				if char == "\n" {

					accomulator = ""
					break
				}
				char_index += len(char)
			}
		case "-d", "-m", "-s":
			t.tokens = append(t.tokens, accomulator)
			accomulator = ""
			// char_index += len(char)

		default:

			switch char {

			case ":", "{", "}", ",":

				if len(accomulator) != 0 && accomulator != " " {
					t.tokens = append(t.tokens, accomulator)
				}
				t.tokens = append(t.tokens, char)
				accomulator = ""

			case " ":
				if len(accomulator) != 0 && accomulator != " " {
					// PrintFl("accomulator: %#v", accomulator)
					t.tokens = append(t.tokens, accomulator)
				}
				accomulator = ""

			default:
				if char != "\n" && char != " " {
					accomulator += char
					// PrintFl("accomulator: %#v", accomulator)
				}
			}
			char_index += len(char)
		}
	}
}
