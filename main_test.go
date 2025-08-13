package main

import "testing"

func TestLunchPhases(t *testing.T) {

	m := InitManager()
	PrintFl("hello")
	m.lunch_training_from_phases("./my_workouts/morning_exercise.tr")
}
