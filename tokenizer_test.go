package main

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	tk := Tokenizer{}
	tk.tokenize_entire_file("./my_workouts/base.txt")
	tk.print_tokens()

}
func TestPullTokens(t *testing.T) {
	tk := Tokenizer{}
	tk.tokenize_entire_file("./my_workouts/base.txt")
	tk.print_tokens()

	for {
		token := tk.PullToken()

		if token == TOKEN_EOF {
			break
		}

		PrintFl("Token: %v", token)
	}

}
