package plugins

import (
	"fmt"
	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

type RiffInvoker struct {
	Application   application.Application
	Handler       string
	Layers        layers.Layers
	InvokerLayer  layers.DependencyLayer
	FunctionLayer layers.Layer
}


// String makes RiffInvoker satisfy the Stringer interface.
func (r RiffInvoker) String() string {
	return fmt.Sprintf("RiffInvoker{ Application: %s, Handler: %s, InvokerLayer: %s, Layers: %s, FunctionLayer: %s }",
		r.Application, r.Handler, r.InvokerLayer, r.Layers, r.FunctionLayer)
}