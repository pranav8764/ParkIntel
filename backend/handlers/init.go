package handlers

import (
	"github.com/pranav8764/ParkIntel/backend/inference"
)

var modelSessions *inference.ModelSessions

// Init initializes the handlers package with the shared ONNX sessions
func Init(sessions *inference.ModelSessions) {
	modelSessions = sessions
}
