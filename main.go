package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed audio_files/*.mp3
var default_audio_sounds embed.FS

type SoundAlias int

const (
	READY_GO      SoundAlias = iota
	WORKOUT       SoundAlias = iota
	REST          SoundAlias = iota
	WORKOUT_ENDED SoundAlias = iota
)

type SoundStream struct {
	Stream beep.StreamSeekCloser
	Format beep.Format
}

type Manager struct {
	DefaultSounds      map[SoundAlias]SoundStream
	UserSounds         map[string]SoundStream
	CurrentSessionName string

	// This approach is chosen so we could easily restart the session from a specific phase and workout in future updates
	CurrentPhaseIndex   int
	CurrentWorkoutIndex int
}

func InitManager() *Manager {

	m := &Manager{
		DefaultSounds: map[SoundAlias]SoundStream{},
		UserSounds:    map[string]SoundStream{},
	}

	m.init_sounds()
	return m

}

func main() {

	workout_session_in_minutes := flag.String("d", "", "How many minutes you like this session last, d=duration")
	workout_time_in_seconds := flag.String("w", "45", "Provide workout time, in seconds, w=workout")
	rest_time_in_seconds := flag.String("r", "0", "Provide rest time between sets, in seconds, if not provided no rest, r=rest")

	flag.Usage = func() {

		fmt.Fprintln(os.Stderr, "trainr have two modes, from file")
		fmt.Fprintln(os.Stderr, "trainer ./[woroutfile]")

		fmt.Fprintln(os.Stderr, "or cmd args, when you want just a quick exercise you can use: ")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Default: d=infinite r=0, w=45 ")
		fmt.Fprintln(os.Stderr, "\nExample:")
		fmt.Fprintln(os.Stdin, "-d 10 -r 15 w 20")

	}

	flag.Parse()

	PrintFl("args: %v", flag.Args())

	m := InitManager()

	if len(flag.Args()) > 0 {
		m.lunch_training_from_phases(flag.Arg(0))
	} else {
		m.lunch_training_from_cmd(workout_session_in_minutes, workout_time_in_seconds, rest_time_in_seconds)
	}

}

func (m *Manager) lunch_training_from_cmd(workout_session_in_minutes, workout_time_in_seconds, rest_time_in_seconds *string) {

	fmt.Printf("rest_time_in_seconds: %v \n", *rest_time_in_seconds)
	fmt.Printf("workout_time_in_seconds: %v \n", *workout_time_in_seconds)
	is_workout_phase := true
	is_init := true

	m.play_sound(m.DefaultSounds[READY_GO])

	end_workout := make(chan struct{})

	if workout_session_in_minutes != nil && *workout_session_in_minutes != "" {

		go func() {

			workout_timer(*workout_session_in_minutes, MINUTES)
			end_workout <- struct{}{}

		}()
	}

	is_workout_ended := false
MAIN_LOOP:
	for {

		select {
		case <-end_workout:
			is_workout_ended = true
		default:

			if is_workout_ended {
				break MAIN_LOOP
			}

			if is_workout_phase {
				// TODO: this is ugly as hell, make it better
				if !is_init {
					WORKOUT_SOUND := m.DefaultSounds[WORKOUT]
					m.play_sound(WORKOUT_SOUND)
				}

				fmt.Printf("WORKOUT phase t: %v \n", time.Now().Format(time.TimeOnly))

				if is_init {
					is_init = false
				}

				workout_timer(*workout_time_in_seconds, SECONDS)
				is_workout_phase = false
			} else {
				if *rest_time_in_seconds != "0" {

					REST_SOUND := m.DefaultSounds[REST]
					m.play_sound(REST_SOUND)
					fmt.Printf("REST phase t: %v \n", time.Now().Format(time.TimeOnly))
					workout_timer(*rest_time_in_seconds, SECONDS)
				}
				is_workout_phase = true
			}
		}
	}

	WORKOUT_ENDED_SOUND := m.DefaultSounds[WORKOUT_ENDED]
	m.play_sound(WORKOUT_ENDED_SOUND)
}

func (m *Manager) lunch_training_from_phases(filepath string) {

	parser := InitParser(filepath, "")
	parser.create_workout_config()

	directory_to_read := parser.current_config.DefaultDumpFolder + parser.current_config.SessionName
	present_audio_files, read_directory_err := os.ReadDir(directory_to_read)

	if read_directory_err != nil {
		PrintFl("[ERROR] read custom audio directory: %v ", read_directory_err.Error())
		log.Fatal()

	}

	for _, file := range present_audio_files {
		file_name := strings.TrimSuffix(file.Name(), ".mp3")
		map_reference := parser.current_config.SessionName + "/" + file_name
		m.UserSounds[map_reference] = sound_to_stream(map_reference, parser.current_config.DefaultDumpFolder)
	}

	m.CurrentSessionName = parser.current_config.SessionName

	current_session := parser.current_config

	phase_traker := make(chan struct{}, 1)
	workout_traker := make(chan struct{}, 1)
	rest_traker := make(chan struct{}, 1)
	end_traker := make(chan struct{}, 1)

	// TODO: propagate all configs to workouts so we can just use a correct config for each workout
	// loop the config "tree" and set
	for _, phase := range current_session.Phases {
		for _, workout := range phase.Workouts {
			// when time is set we do not edit it
			if workout.TimeWorkout == "" {
				if phase.TimeWorkout != "" {
					workout.TimeWorkout = phase.TimeWorkout
				} else {
					workout.TimeWorkout = current_session.TimeWorkout
				}
			}

			if workout.TimeRest == "" {
				if phase.TimeRest != "" {
					workout.TimeRest = phase.TimeRest
				} else {
					workout.TimeRest = current_session.TimeRest
				}
			}

		}
	}

MAIN_SESSION_LOOP:
	for {

		select {

		case <-phase_traker:

			phase := current_session.Phases[m.CurrentPhaseIndex]
			m.play_sound_current_session(phase.PhaseName)
			PrintFl("phase_traker %v", phase.PhaseName)

			//just to postpone the workout name sound start from the session name
			time.Sleep(time.Second * 2)

			workout_traker <- struct{}{}

		case <-workout_traker:

			phase := current_session.Phases[m.CurrentPhaseIndex]
			workout := phase.Workouts[m.CurrentWorkoutIndex]

			m.play_sound_current_session(workout.WorkoutName)
			workout_timer_any_accepted_time(workout.TimeWorkout)

			//TODO: this can be factor to a func
			if m.CurrentWorkoutIndex < len(phase.Workouts)-1 {

				if workout.TimeRest != "" {
					rest_traker <- struct{}{}
				} else {
					m.CurrentWorkoutIndex++
					workout_traker <- struct{}{}
				}

			} else {
				if m.CurrentPhaseIndex < len(current_session.Phases)-1 {
					m.CurrentPhaseIndex++
					m.CurrentWorkoutIndex = 0
					phase_traker <- struct{}{}
				} else {
					end_traker <- struct{}{}
				}
			}

		case <-rest_traker:

			phase := current_session.Phases[m.CurrentPhaseIndex]
			workout := phase.Workouts[m.CurrentWorkoutIndex]

			m.play_sound(m.DefaultSounds[REST])
			workout_timer_any_accepted_time(workout.TimeRest)
			m.CurrentWorkoutIndex++
			workout_traker <- struct{}{}

		case <-end_traker:

			WORKOUT_ENDED_SOUND := m.DefaultSounds[WORKOUT_ENDED]
			m.play_sound(WORKOUT_ENDED_SOUND)
			break MAIN_SESSION_LOOP

		default:
			m.play_sound(m.DefaultSounds[READY_GO])
			phase_traker <- struct{}{}

		}
	}
}

func sound_to_stream(sound_name, base_folder string) SoundStream {
	workout_file, err := os.Open(base_folder + sound_name + ".mp3")
	if err != nil {
		log.Fatal(err)
	}
	streamer, format, err := mp3.Decode(workout_file)
	if err != nil {
		log.Fatal(err)
	}

	return SoundStream{
		Stream: streamer,
		Format: format,
	}
}

// TODO:  Will be enabled later
// func sound_from_embeded(sound_name string) SoundStream {
//
//		workout_file, err := default_audio_sounds.Open("sounds/" + sound_name + ".mp3")
//
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		streamer, format, err := mp3.Decode(workout_file)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		return SoundStream{
//			Stream: streamer,
//			Format: format,
//		}
//	}

func (m *Manager) play_sound_current_session(file_name string) {

	m.play_sound(m.UserSounds[m.CurrentSessionName+"/"+file_name])
}

func (m *Manager) init_sounds() {
	// TODO: this should work for final build
	// m.DefaultSounds[REST] = sound_from_embeded("rest")
	// m.DefaultSounds[WORKOUT] = sound_from_embeded("workout")
	// m.DefaultSounds[READY_GO] = sound_from_embeded("ready_go")
	// m.DefaultSounds[WORKOUT_ENDED] = sound_from_embeded("end_workout")

	// TODO: imbed this default sounds in to the binary

	// old way
	default_sounds_folder := "./audio_files/"
	m.DefaultSounds[REST] = sound_to_stream("rest", default_sounds_folder)
	m.DefaultSounds[WORKOUT] = sound_to_stream("workout", default_sounds_folder)
	m.DefaultSounds[READY_GO] = sound_to_stream("ready_go", default_sounds_folder)
	m.DefaultSounds[WORKOUT_ENDED] = sound_to_stream("end_workout", default_sounds_folder)
}

func (m *Manager) play_sound(sound SoundStream) {

	speaker.Init(sound.Format.SampleRate, sound.Format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(sound.Stream, beep.Callback(func() {
		sound.Stream.Seek(0)
		done <- true
	})))
	<-done
}

type TimeDuration string

const (
	MINUTES TimeDuration = "m"
	SECONDS TimeDuration = "s"
)

func workout_timer(time_in_seconds string, duration_to_parce TimeDuration) {
	workout_time_in_seconds, duration_parse_err := time.ParseDuration(time_in_seconds + string(duration_to_parce))
	if duration_parse_err != nil {
		fmt.Printf("Cannot parse workout time")
		log.Fatal("")
	}
	time.Sleep(workout_time_in_seconds)
}
func workout_timer_any_accepted_time(time_to_convert string) {
	workout_timer, duration_parse_err := time.ParseDuration(time_to_convert)
	if duration_parse_err != nil {
		fmt.Printf("[Error] Cannot parse workout time %v", duration_parse_err)
		log.Fatal("")
	}
	PrintFl("Duration: %v", workout_timer)
	time.Sleep(workout_timer)
}

func PrintFl(format string, a ...any) {
	fmt.Println()
	_, f_name, f_line, _ := runtime.Caller(1)
	fmt.Printf("%v:%v \n", f_name, f_line)
	fmt.Printf(format, a...)
	fmt.Println()
}
func TODO(format ...any) {
	fmt.Println()
	_, f_name, f_line, _ := runtime.Caller(1)
	fmt.Printf("%v:%v \n", f_name, f_line)
	fmt.Printf("%s", format...)
	fmt.Println()
	log.Fatal()
}
