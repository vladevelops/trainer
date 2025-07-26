package main

import (
	"fmt"
	"log"
	"strings"
)

type Parser struct {
	t *Tokenizer
}

func InitParser(file_path string) *Parser {

	tokenizer := Tokenizer{}
	tokenizer.tokenize_entire_file(file_path)

	p := Parser{
		t: &tokenizer,
	}
	return &p
}

func (p *Parser) get_and_expect_token(expected_token string) error {

	pulled_token := p.t.PullToken()
	if expected_token != pulled_token {
		return fmt.Errorf("expected_token: %v got: %v", expected_token, pulled_token)
	}
	return nil
}

// config tokens
const (
	CONFIG_DURATION = "-d"
	CONFIG_REST     = "-r"
	CONFIG_WORKOUT  = "-w"
)
const (
	PUNCT_COLON      = ":"
	PUNCT_OPEN_BRACE = "{"
)

type WorkOutConfig struct {
	TimeDuration string
	TimeREST     string
	TimeWORKOUT  string
}

func (p *Parser) create_workout_config() {

	new_workout_config := WorkOutConfig{
		TimeDuration: "",
		TimeREST:     "",
		TimeWORKOUT:  "",
	}
	if p.get_and_expect_token(string(PHASES)) == nil {
		if err := p.get_and_expect_token(PUNCT_COLON); err != nil {
			PrintFl("ERROR: %v", err.Error())
			log.Fatal()
		}

		for range 3 {

			config_token := p.t.CheckCurentToken()
			config_value := p.t.PullToken()

			if strings.HasPrefix(config_value, "-") {
				PrintFl("ERROR: config value cannot have -prefix, got: %v", config_value)
				log.Fatal()
			}

			// TODO: check config value to have m|s at the end and reject other strings
			switch config_token {
			case CONFIG_DURATION:
				new_workout_config.TimeDuration = config_value
				continue
			case CONFIG_REST:
				new_workout_config.TimeREST = config_value
				continue
			case CONFIG_WORKOUT:
				new_workout_config.TimeWORKOUT = config_value
				continue
			default:
				PrintFl("ERROR: after colon in PHASES mode we cannot get: %v", config_token)
				log.Fatal()
			}
		}

		// TODO: this is ugly, use reflect to create checking in a loop
		if new_workout_config.TimeDuration == "" {
			PrintFl("ERROR: -d must be set in PHASES but its empty")
			log.Fatal()
		}

		if new_workout_config.TimeREST == "" {
			PrintFl("ERROR: -r must be set in PHASES but its empty")
			log.Fatal()
		}

		if new_workout_config.TimeWORKOUT == "" {
			PrintFl("ERROR: -w must be set in PHASES but its empty")
			log.Fatal()
		}

	} else {
		TODO("create_workout_config no PHASES at the begining")
	}

}
