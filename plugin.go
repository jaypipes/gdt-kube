// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"github.com/jaypipes/gdt"
	gdttypes "github.com/jaypipes/gdt-core/types"
	"gopkg.in/yaml.v3"
)

func init() {
	gdt.RegisterPlugin(Plugin())
}

const (
	pluginName = "kube"
)

type plugin struct{}

func (p *plugin) Info() gdttypes.PluginInfo {
	return gdttypes.PluginInfo{
		Name: pluginName,
	}
}

func (p *plugin) Defaults() yaml.Unmarshaler {
	return &Defaults{}
}

func (p *plugin) Specs() []gdttypes.Spec {
	return []gdttypes.Spec{&Spec{}}
}

// Plugin returns the Kubernetes gdt plugin
func Plugin() gdttypes.Plugin {
	return &plugin{}
}
