package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/go-tts/tts/pkg/speech"
)

type WorkoutConfigs struct {
	TimeRest    string
	TimeWorkout string
	// NOTE: not a fun of this config being here
	TimePhaseRest string
}

type SingleWorkout struct {
	WorkoutName string
	WorkoutConfigs
}

type Phase struct {
	PhaseName string
	Workouts  []*SingleWorkout
	WorkoutConfigs
}

type WorkoutConfig struct {
	Phases            []*Phase
	SessionName       string
	DefaultDumpFolder string
	WorkoutConfigs
}

type Parser struct {
	t              *Tokenizer
	current_config *WorkoutConfig
}

func InitParser(file_path, dump_folder_overwrite string) *Parser {

	tokenizer := Tokenizer{}
	tokenizer.tokenize_entire_file(file_path)

	p := Parser{
		t: &tokenizer,
		current_config: &WorkoutConfig{
			WorkoutConfigs: WorkoutConfigs{
				TimeRest:      "",
				TimeWorkout:   "",
				TimePhaseRest: "",
			},
			DefaultDumpFolder: "./trainer_files/",
			SessionName:       "",
			Phases:            []*Phase{},
		},
	}

	if dump_folder_overwrite != "" {
		p.current_config.DefaultDumpFolder = dump_folder_overwrite

	}
	return &p
}

// config tokens
const (
	CONFIG_REST       = "-r"
	CONFIG_PHASE_REST = "-pr"
	CONFIG_WORKOUT    = "-w"
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
	workout_name := p.t.PullToken()
	p.current_config.SessionName = workout_name
	if err := p.get_and_expect_token(PUNCT_COLON); err != nil {
		PrintFl("ERROR: %v", err.Error())
		return
	}
	p.parse_phases_config()
	for {
		p.parse_single_phase()

		switch p.t.CheckCurentToken() {
		case PUNCT_CLOSE_BRACE:
			p.create_audio_files_from_parsed_config()

			return
		default:
			// TODO: check for punctuation and similar, only a variable name is ok
			continue
		}
	}
}

func (p *Parser) create_audio_files_from_parsed_config() {

	// PrintFl("Parsed Config: %#v ", p.current_config.Phases)
	session_name := p.current_config.SessionName
	folder_name := p.current_config.DefaultDumpFolder + session_name + "/"

	_, folder_present_err := os.Stat(folder_name)
	if os.IsNotExist(folder_present_err) {
		err := os.MkdirAll(folder_name, 0755)
		if err != nil {
			fmt.Println("Error creating folder:", err)
			log.Panic()
		}
	}

	// TODO: clean out unused sounds, get file list and compare it
	// show the generation string only when it really needed

	PrintFl("Generating all the audio files, please wait....")
	session_name_file_path := folder_name + session_name + ".mp3"
	if _, is_existst_err := os.Stat(session_name_file_path); os.IsNotExist(is_existst_err) {
		p.generate_audio_file("Starting.. "+session_name+" session ", session_name_file_path)
	}

	for _, phase := range p.current_config.Phases {
		phase_name_file_path := folder_name + phase.PhaseName + ".mp3"
		if _, is_existst_err := os.Stat(phase_name_file_path); os.IsNotExist(is_existst_err) {
			p.generate_audio_file("Starting.. "+phase.PhaseName+" phase ", phase_name_file_path)
		}

		for _, workout := range phase.Workouts {
			workout_name_file_path := folder_name + workout.WorkoutName + ".mp3"
			if _, is_existst_err := os.Stat(workout_name_file_path); os.IsNotExist(is_existst_err) {
				p.generate_audio_file(workout.WorkoutName+" workout. ", workout_name_file_path)
			}
		}
	}
	PrintFl("Generating audio files done")

}

func (p *Parser) generate_audio_file(text_to_voice, output_path string) {

	output, _ := os.Create(output_path)
	if err := speech.WriteToAudioStream(strings.NewReader(text_to_voice), output, "en"); err != nil {
		PrintFl("generate audio: %v ", err.Error())
		log.Fatal()
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
		parsed_workout := p.must_parse_single_workout()
		phase.Workouts = append(phase.Workouts, &parsed_workout)
		comma_or_close_brase := p.t.CheckCurentToken()

		switch comma_or_close_brase {
		case PUNCT_COMMA:
			p.t.PullToken()

			comma_or_close_brase := p.t.CheckCurentToken()

			if comma_or_close_brase == PUNCT_CLOSE_BRACE {
				break SINGLE_WORKOUT_AFTER_CONFIG
			}
			continue
		case PUNCT_CLOSE_BRACE:
			break SINGLE_WORKOUT_AFTER_CONFIG
		}
	}
}

func (p *Parser) must_parse_single_workout() (workout SingleWorkout) {

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
			panic("")
		}

		config_durations := p.parse_overwriteable_config()
		workout.TimeRest = config_durations.TimeRest
		workout.TimeWorkout = config_durations.TimeWorkout

		p.t.parser_position -= 1
		return workout
	default:
		PrintFl("ERROR:parse_single_workout: expecting PUNCT_COMMA or PUNCT_COLON got %v", colon_or_comma)
		panic("")
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

	AMOUNT_OF_MANDATORY_CONFIGS := reflect.ValueOf(WorkoutConfigs{}).NumField()

	for range AMOUNT_OF_MANDATORY_CONFIGS {

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
		case CONFIG_REST:
			p.current_config.TimeRest = config_value
			continue
		case CONFIG_PHASE_REST:
			p.current_config.TimePhaseRest = config_value
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

func (p *Parser) parse_overwriteable_config() (config_durations WorkoutConfigs) {

	for {

		config_token := p.t.CheckCurentToken()
		// PrintFl("config_token %+v \n", config_token)

		// here we are getting the config token,
		// but we have no idea witch so we cannot use `get_and_expect_token`
		p.t.PullToken()

		config_value := p.t.PullToken()
		// PrintFl("config_value %+v \n", config_value)

		if strings.HasPrefix(config_value, "-") {
			PrintFl("ERROR: config value cannot have `-` prefix, got: %v", config_value)
			log.Fatal()
		}

		if config_token != PUNCT_OPEN_BRACE && config_token != PUNCT_COMMA && !p.check_config_value_is_valid(config_value) {
			PrintFl("ERROR: config value can be only in s[econds] | m[inutes] got: %v", config_value)
			panic("")
			// log.Fatal()
		}

		switch config_token {
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
