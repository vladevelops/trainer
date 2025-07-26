package main

import "testing"

func TestParseConfig(t *testing.T) {

	parser := InitParser("./my_workouts/base.txt")
	parser.create_workout_config()
}
