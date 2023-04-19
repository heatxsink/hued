package presets

import (
	"github.com/heatxsink/go-hue/lights"
)

var (
	onState            lights.State
	offState           lights.State
	redState           lights.State
	blueState          lights.State
	energizeState      lights.State
	relaxState         lights.State
	readingState       lights.State
	concentrateState   lights.State
	candleLightState   lights.State
	virginAmericaState lights.State
	whiteState         lights.State
	orangeState        lights.State
	deepSeaState       lights.State
	greenState         lights.State
	snowState          lights.State
	movieModeState     lights.State
	buttonStates       []string
)

func init() {
	onState = lights.State{On: true}
	offState = lights.State{On: false}
	redState = lights.State{On: true, Hue: 65527, Effect: "none", Bri: 13, Sat: 253, CT: 500, XY: []float32{0.6736, 0.3221}, Alert: "none", TransitionTime: 4}
	blueState = lights.State{On: true, Hue: 46573, Effect: "none", Bri: 254, Sat: 251, CT: 500, XY: []float32{0.1754, 0.0556}, Alert: "none", TransitionTime: 4}
	energizeState = lights.State{On: true, Hue: 34495, Effect: "none", Bri: 254, Sat: 232, CT: 155, XY: []float32{0.3151, 0.3252}, Alert: "none", TransitionTime: 4}
	relaxState = lights.State{On: true, Hue: 13088, Effect: "none", Bri: 144, Sat: 212, CT: 467, XY: []float32{0.5128, 0.4147}, Alert: "none", TransitionTime: 4}
	readingState = lights.State{On: true, Hue: 15331, Effect: "none", Bri: 222, Sat: 121, CT: 343, XY: []float32{0.4448, 0.4066}, Alert: "none", TransitionTime: 4}
	concentrateState = lights.State{On: true, Hue: 33849, Effect: "none", Bri: 254, Sat: 44, CT: 234, XY: []float32{0.3693, 0.3695}, Alert: "none", TransitionTime: 4}
	candleLightState = lights.State{On: true, Hue: 15339, Effect: "none", Bri: 19, Sat: 120, CT: 343, XY: []float32{0.4443, 0.4064}, Alert: "none", TransitionTime: 4}
	virginAmericaState = lights.State{On: true, Hue: 54179, Effect: "none", Bri: 254, Sat: 253, CT: 223, XY: []float32{0.3621, 0.1491}, Alert: "none", TransitionTime: 4}
	whiteState = lights.State{On: true, Hue: 34495, Effect: "none", Bri: 203, Sat: 232, CT: 155, XY: []float32{0.3151, 0.3252}, Alert: "none", TransitionTime: 4}
	orangeState = lights.State{On: true, Hue: 4868, Effect: "none", Bri: 254, Sat: 252, CT: 500, XY: []float32{0.6225, 0.3594}, Alert: "none", TransitionTime: 4}
	deepSeaState = lights.State{On: true, Hue: 65527, Effect: "none", Bri: 253, Sat: 253, CT: 500, XY: []float32{0.6736, 0.3221}, Alert: "none", TransitionTime: 0}
	greenState = lights.State{On: true, Hue: 25654, Effect: "none", Bri: 254, Sat: 253, CT: 290, XY: []float32{0.4083, 0.5162}, Alert: "none", TransitionTime: 4}
	snowState = lights.State{On: true, Hue: 34258, Effect: "none", Bri: 254, Sat: 176, CT: 181, XY: []float32{0.3327, 0.3413}, Alert: "none", TransitionTime: 4}
	movieModeState = lights.State{On: true, Hue: 65527, Effect: "none", Bri: 51, Sat: 253, CT: 500, XY: []float32{0.6736, 0.3221}, Alert: "none", TransitionTime: 4}
	buttonStates = []string{"deep-sea", "blue", "relax", "reading", "concentrate", "candle-light"}
}

func GetButtonStates() []string {
	return buttonStates
}

func GroupName(name string) int {
	returnValue := -1
	if name == "all" {
		returnValue = 0
	} else if name == "bedroom" {
		returnValue = 1
	} else if name == "living-room" {
		returnValue = 2
	} else if name == "hallway" {
		returnValue = 3
	} else if name == "master-bedroom" {
		returnValue = 4
	} else if name == "kitchen" {
		returnValue = 5
	}
	return returnValue
}

func GetLightState(name string) lights.State {
	switch name {
	case "on":
		return onState
	case "off":
		return offState
	case "red":
		return redState
	case "blue":
		return blueState
	case "energize":
		return energizeState
	case "relax":
		return relaxState
	case "reading":
		return readingState
	case "concentrate":
		return concentrateState
	case "candle-light":
		return candleLightState
	case "virgin-america":
		return virginAmericaState
	case "white":
		return whiteState
	case "orange":
		return orangeState
	case "deep-sea":
		return deepSeaState
	case "green":
		return greenState
	case "snow":
		return snowState
	case "movie-mode":
		return movieModeState
	default:
		return onState
	}
}
