build: 
	@go build -o trainer .

file: 
	@go run . ./my_workouts/morning_exercise.tr

