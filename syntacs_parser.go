package main

import (
	"fmt"
	"log"
	"strings"
)

type WorkoutDurationsConfigs struct {
	TimeDuration string
	TimeRest     string
	TimeWorkout  string
}
type SingleWorkout struct {
	WorkoutName string
	WorkoutDurationsConfigs
}

type Phase struct {
	PhaseName string
	Workouts  []*SingleWorkout
	WorkoutDurationsConfigs
}

type WorkoutConfig struct {
	Phases []*Phase
	WorkoutDurationsConfigs
}

type Parser struct {
	t              *Tokenizer
	current_config *WorkoutConfig
}

func InitParser(file_path string) *Parser {

	tokenizer := Tokenizer{}
	tokenizer.tokenize_entire_file(file_path)

	p := Parser{
		t: &tokenizer,
		current_config: &WorkoutConfig{
			WorkoutDurationsConfigs: WorkoutDurationsConfigs{
				TimeDuration: "",
				TimeRest:     "",
				TimeWorkout:  "",
			},
			Phases: []*Phase{},
		},
	}

	return &p
}

// config tokens
const (
	CONFIG_DURATION = "-d"
	CONFIG_REST     = "-r"
	CONFIG_WORKOUT  = "-w"
)

// config punctuation
const (
	PUNCT_COLON       = ":"
	PUNCT_COMMA       = ","
	PUNCT_OPEN_BRACE  = "{"
	PUNCT_CLOSE_BRACE = "}"
)

// config session time
const (
	CONFIG_SECONDS = "s"
	CONFIG_MINUTES = "m"
)

func (p *Parser) create_workout_config() {

	if p.get_and_expect_token(string(PHASES)) == nil {
		if err := p.get_and_expect_token(PUNCT_COLON); err != nil {
			PrintFl("ERROR: %v", err.Error())
			return
		}
		p.parse_phases_config()
		for {
			p.parse_single_phase()

			switch p.t.CheckCurentToken() {
			case PUNCT_CLOSE_BRACE:

				// PrintFl("Parsed Config: %#v ", p.current_config.Phases)
				for _, phase := range p.current_config.Phases {
					PrintFl("phase: %+v", phase.PhaseName)
					for _, workout := range phase.Workouts {

						PrintFl("workout: %+v", workout)

					}

				}

				return
			default:
				// TODO: check for punctuation and similar, only a variable name is ok
				continue
			}

		}

	} else {

		TODO("create_workout_config no PHASES at the begining")
	}
}

func (p *Parser) parse_single_phase() {

	phase := Phase{}

	phase_name := p.t.PullToken()

	// TODO: function to check if the token is some sort of a config, punctuation, etc..
	phase.PhaseName = phase_name
	// TODO: expect COLON and config?
	if err := p.get_and_expect_token(PUNCT_COLON); err != nil {
		PrintFl("ERROR: %v", err.Error())
		return
	}
	config_or_open_brace := p.t.CheckCurentToken()

	switch config_or_open_brace {
	case PUNCT_OPEN_BRACE:

		if err := p.get_and_expect_token(PUNCT_OPEN_BRACE); err != nil {
			PrintFl("ERROR: %v", err.Error())
			return
		}

		p.parse_multiple_workouts(&phase)
		p.current_config.Phases = append(p.current_config.Phases, &phase)
		p.t.PullToken()

	default:
		config_durations := p.parse_overwriteable_config()
		phase.TimeDuration = config_durations.TimeDuration
		phase.TimeRest = config_durations.TimeRest
		phase.TimeWorkout = config_durations.TimeWorkout
		p.parse_multiple_workouts(&phase)
		p.current_config.Phases = append(p.current_config.Phases, &phase)
		p.t.PullToken()
	}
}

func (p *Parser) parse_multiple_workouts(phase *Phase) {

SINGLE_WORKOUT_AFTER_CONFIG:
	for {
		parsed_workout := p.parse_single_workout()
		phase.Workouts = append(phase.Workouts, &parsed_workout)
		comma_or_close_brase := p.t.CheckCurentToken()

		switch comma_or_close_brase {
		case PUNCT_COMMA:
			p.t.PullToken()
			continue
		case PUNCT_CLOSE_BRACE:
			break SINGLE_WORKOUT_AFTER_CONFIG
		}
	}
}

func (p *Parser) parse_single_workout() (workout SingleWorkout) {

	workout_name := p.t.PullToken()

	// PrintFl("workout_name: %v ", workout_name)
	// log.Fatal()
	workout.WorkoutName = workout_name

	colon_or_comma := p.t.CheckCurentToken()

	switch colon_or_comma {
	case PUNCT_CLOSE_BRACE:
		return workout

	case PUNCT_COMMA:
		return workout
	case PUNCT_COLON:

		if err := p.get_and_expect_token(PUNCT_COLON); err != nil {
			PrintFl("ERROR: %v", err)
			return
		}

		config_durations := p.parse_overwriteable_config()
		workout.TimeDuration = config_durations.TimeDuration
		workout.TimeRest = config_durations.TimeRest
		workout.TimeWorkout = config_durations.TimeWorkout

		p.t.parser_position -= 1
		return workout
	default:
		PrintFl("ERROR:parse_single_workout: expecting PUNCT_COMMA or PUNCT_COLON got %v", colon_or_comma)
		return
	}
}

func (p *Parser) get_and_expect_token(expected_token string) error {

	pulled_token := p.t.PullToken()
	if expected_token != pulled_token {
		return fmt.Errorf("expected_token: %v got: %v", expected_token, pulled_token)
	}
	return nil
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
			p.current_config.TimeRest = config_value
			continue
		case CONFIG_WORKOUT:
			p.current_config.TimeWorkout = config_value
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

	if p.current_config.TimeRest == "" {
		PrintFl("ERROR: -r must be set in PHASES but its empty")
		log.Fatal()
	}

	if p.current_config.TimeWorkout == "" {
		PrintFl("ERROR: -w must be set in PHASES but its empty")
		log.Fatal()
	}

	if err := p.get_and_expect_token(PUNCT_OPEN_BRACE); err != nil {
		PrintFl("ERROR: %v", err.Error())
		return
	}
}

func (p *Parser) parse_overwriteable_config() (config_durations WorkoutDurationsConfigs) {

	for {

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
			config_durations.TimeDuration = config_value
			continue
		case CONFIG_REST:
			config_durations.TimeRest = config_value
			continue
		case CONFIG_WORKOUT:
			config_durations.TimeWorkout = config_value
			continue
		case PUNCT_OPEN_BRACE:

			p.t.parser_position -= 1

			return config_durations
		case PUNCT_COMMA:
			p.t.parser_position -= 1
			return config_durations
		default:
			PrintFl("ERROR: after colon in OWERWRITE mode we cannot get: %v", config_token)
			log.Fatal()
		}
	}
}

func (p *Parser) check_config_value_is_valid(config_value string) bool {

	last_string_in_config_value := string(config_value[len(config_value)-1])
	if last_string_in_config_value == CONFIG_SECONDS ||
		last_string_in_config_value == CONFIG_MINUTES {

		return true
	}
	return false

}
