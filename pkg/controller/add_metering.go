package controller

import (
	"github.com/cs-operators/metering-operator/pkg/controller/metering"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, metering.Add)
}
