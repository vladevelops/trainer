
> [!NOTE]  
> This project is new and may have bugs, feel free to report them.
> the code needs cleaning, as soon as i have some more free time it will be done

# Your Trainer App - Workout Timer with Custom Audio


## Why?
I created this project for myself to get back in shape, and i want to share it with everyone.
As we all know, sitting for long hours in front of a computer can become a issue for your health, and we often forget to exercise.
This app lets you easily start a workout right from your desk without the annoying ads you find in other free fitness apps.
It's designed to be simple, and allow you to integrate exercise routines into your day

## Features

- **Custom sound**: Generates audio from the provided workout names.
- **Customizable workout sessions**: Configure sessions with different phases, workouts, and rest times.
- **Flexible time management**: Specify a duration for each workout and rest period.
- **Support for multiple phases**: A workout session can consist of multiple phases, each with its own set of workouts and rest times.


## Installation

### Clone the repository

```bash
git clone https://github.com/yourusername/trainer-app.git
cd trainer-app
go build .
```



### Simple language for creating exercises

```bash

Exercise example:

morning_exercise: -w 45s -r 15s -pr 1m {
  worm-up: -w 1m -r 5s { # owerwrides
    jumping-open-arms, 
    elbow-to-knees,  
    squats-elbow,
    jumping-open-arms,
  }
  standing-workout: {
    weights-hands-closed,
    weights-hands-up,
    squats-weights,
  }
  ground: -w 50s {
    weights-hands-up,
    plank,
    push-ups,
    swimmer,
  }
  cool-down: -w 60s {
    jumping-open-arms,
    stretching,
  }
}

# configs:
# -r  == rest
# -w  == workout
# -pr == phase rest (time) -> only for WORKOUT_SESSION_CONFIGS 
#
#
#  m == minutes
#  s == seconds
#  
# `FOLDER_NAME(variable)`: ...(WORKOUT_SESSION_CONFIGS)configs {
#   `PHASE_NAME(variable)`: [...configs] {
#     ...`SINGLE_WORKOUT_NAME(variable)`: [...configs]
#   }
# }

# FOLDER_NAME         -> will be the folder where the generated audio will be outed, the configs here are mandatory and are the default
# PHASE_NAME          -> name for each phase of the exercise session, you can overwrite the configs for each phase
# SINGLE_WORKOUT_NAME -> single workout name, you can overwrite the configs for each workout

# NOTE: the go-tts pronounces `_` in word but `-` is treated like small pause in speech
# if you want to add more pause between words you can use `...` method, more dots more the pause is long

```

All names are assign by you, i personally do not know all the fancy term for all the workouts so i call them as i want.

## Future Developments
 - TUI interface for ease of interaction
 - UI interface, so it can be ported to a phone application for example

