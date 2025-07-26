package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

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
	Sounds map[SoundAlias]SoundStream
}

func sound_to_stream(sound_name string) SoundStream {

	workout_file, err := os.Open("./audio_files/" + sound_name + ".mp3")
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

func InitManager() *Manager {

	m := &Manager{
		Sounds: map[SoundAlias]SoundStream{},
	}
	m.init_sounds()
	return m

}

func (m *Manager) init_sounds() {
	// NOTE: we can create more flexible way to get the file names form dir
	m.Sounds[REST] = sound_to_stream("rest")
	m.Sounds[WORKOUT] = sound_to_stream("workout")
	m.Sounds[READY_GO] = sound_to_stream("ready_go")
	m.Sounds[WORKOUT_ENDED] = sound_to_stream("end_workout")
}

func (m *Manager) play_sound(sound_alias SoundAlias) {
	sound := m.Sounds[sound_alias]

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

func main() {

	rest_time_in_seconds := flag.String("r", "0", "Provide rest time between sets, in seconds, if not provided no rest, r=rest")
	workout_time_in_seconds := flag.String("w", "45", "Provide workout time, in seconds, w=workout")
	workout_session_in_minutes := flag.String("d", "", "How many minutes you like this session last, d=duration")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Default: d=infinite r=0, w=45 ")
		flag.PrintDefaults()

		fmt.Fprintln(os.Stderr, "\nExample:")
		fmt.Fprintln(os.Stdin, "-d 10 -r 15 w 20")

	}

	flag.Parse()

	fmt.Printf("rest_time_in_seconds: %v \n", *rest_time_in_seconds)
	fmt.Printf("workout_time_in_seconds: %v \n", *workout_time_in_seconds)

	m := InitManager()

	is_workout_phase := true
	is_init := true
	m.play_sound(READY_GO)

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
					m.play_sound(WORKOUT)
				}

				fmt.Printf("WORKOUT phase t: %v \n", time.Now().Format(time.TimeOnly))

				if is_init {
					is_init = false
				}

				workout_timer(*workout_time_in_seconds, SECONDS)
				is_workout_phase = false
			} else {
				if *rest_time_in_seconds != "0" {
					m.play_sound(REST)
					fmt.Printf("REST phase t: %v \n", time.Now().Format(time.TimeOnly))
					workout_timer(*rest_time_in_seconds, SECONDS)
				}
				is_workout_phase = true
			}
		}
	}

	m.play_sound(WORKOUT_ENDED)
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
