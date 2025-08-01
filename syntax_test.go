package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {

	parser := InitParser("./my_workouts/base.tr", "")
	parser.create_workout_config()
}
