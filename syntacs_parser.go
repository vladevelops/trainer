package main

import (
	"fmt"
	"log"
	"strings"
)

type WorkOutConfig struct {
	TimeDuration string
	TimeREST     string
	TimeWORKOUT  string
}

type Parser struct {
	t              *Tokenizer
	current_config *WorkOutConfig
}

func InitParser(file_path string) *Parser {

	tokenizer := Tokenizer{}
	tokenizer.tokenize_entire_file(file_path)

	p := Parser{
		t: &tokenizer,
		current_config: &WorkOutConfig{
			TimeDuration: "",
			TimeREST:     "",
			TimeWORKOUT:  "",
		},
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

const (
	CONFIG_SECONDS = "s"
	CONFIG_MINUTES = "m"
)

func (p *Parser) check_config_value_is_valid(config_value string) bool {

	last_string_in_config_value := string(config_value[len(config_value)-1])
	if last_string_in_config_value == CONFIG_SECONDS ||
		last_string_in_config_value == CONFIG_MINUTES {

		return true
	}
	return false

}
func (p *Parser) parse_phases_config() {

	for range 3 {

		config_token := p.t.CheckCurentToken()

		// here we are getting the config token,
		// but we have no idea witch so we cannot use `get_and_expect_token`
		p.t.PullToken()

		config_value := p.t.PullToken()

		if strings.HasPrefix(config_value, "-") {
			PrintFl("ERROR: config value cannot have `-` prefix, got: %v", config_value)
			log.Fatal()
		}

		if !p.check_config_value_is_valid(config_value) {
			PrintFl("ERROR: config value can be only in s[econds] | m[inutes] got: %v", config_value)
			log.Fatal()
		}

		switch config_token {
		case CONFIG_DURATION:
			p.current_config.TimeDuration = config_value

			continue
		case CONFIG_REST:
			p.current_config.TimeREST = config_value
			continue
		case CONFIG_WORKOUT:
			p.current_config.TimeWORKOUT = config_value
			continue
		default:
			PrintFl("ERROR: after colon in PHASES mode we cannot get: %v", config_token)
			log.Fatal()
		}
	}

	// TODO: this is ugly, use reflect to create checking in a loop
	if p.current_config.TimeDuration == "" {
		PrintFl("ERROR: -d must be set in PHASES but its empty")
		log.Fatal()
	}

	if p.current_config.TimeREST == "" {
		PrintFl("ERROR: -r must be set in PHASES but its empty")
		log.Fatal()
	}

	if p.current_config.TimeWORKOUT == "" {
		PrintFl("ERROR: -w must be set in PHASES but its empty")
		log.Fatal()
	}
}

func (p *Parser) create_workout_config() {

	if p.get_and_expect_token(string(PHASES)) == nil {
		if err := p.get_and_expect_token(PUNCT_COLON); err != nil {
			PrintFl("ERROR: %v", err.Error())
			log.Fatal()
		}
		p.parse_phases_config()
		PrintFl("Parsed Config: %#v ", p.current_config)

	} else {
		TODO("create_workout_config no PHASES at the begining")
	}
}
