// generated by go-extpoints -- DO NOT EDIT
package extpoints

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
)

var extRegistry = &registryType{m: make(map[string]*extensionPoint)}

type registryType struct {
	sync.Mutex
	m map[string]*extensionPoint
}

// Top level registration

func extensionTypes(extension interface{}) []string {
	var ifaces []string
	typ := reflect.TypeOf(extension)
	for name, ep := range extRegistry.m {
		if ep.iface.Kind() == reflect.Func && typ.AssignableTo(ep.iface) {
			ifaces = append(ifaces, name)
		}
		if ep.iface.Kind() != reflect.Func && typ.Implements(ep.iface) {
			ifaces = append(ifaces, name)
		}
	}
	return ifaces
}

func RegisterExtension(extension interface{}, name string) []string {
	extRegistry.Lock()
	defer extRegistry.Unlock()
	var ifaces []string
	for _, iface := range extensionTypes(extension) {
		if extRegistry.m[iface].register(extension, name) {
			ifaces = append(ifaces, iface)
		}
	}
	return ifaces
}

func UnregisterExtension(name string) []string {
	extRegistry.Lock()
	defer extRegistry.Unlock()
	var ifaces []string
	for iface, extpoint := range extRegistry.m {
		if extpoint.unregister(name) {
			ifaces = append(ifaces, iface)
		}
	}
	return ifaces
}

// Base extension point

type extensionPoint struct {
	sync.Mutex
	iface      reflect.Type
	extensions map[string]interface{}
}

func newExtensionPoint(iface interface{}) *extensionPoint {
	ep := &extensionPoint{
		iface:      reflect.TypeOf(iface).Elem(),
		extensions: make(map[string]interface{}),
	}
	extRegistry.Lock()
	extRegistry.m[ep.iface.Name()] = ep
	extRegistry.Unlock()
	return ep
}

func (ep *extensionPoint) lookup(name string) interface{} {
	ep.Lock()
	defer ep.Unlock()
	ext, ok := ep.extensions[name]
	if !ok {
		return nil
	}
	return ext
}

func (ep *extensionPoint) all() map[string]interface{} {
	ep.Lock()
	defer ep.Unlock()
	all := make(map[string]interface{})
	for k, v := range ep.extensions {
		all[k] = v
	}
	return all
}

func (ep *extensionPoint) register(extension interface{}, name string) bool {
	ep.Lock()
	defer ep.Unlock()
	if name == "" {
		typ := reflect.TypeOf(extension)
		if typ.Kind() == reflect.Func {
			nameParts := strings.Split(runtime.FuncForPC(
				reflect.ValueOf(extension).Pointer()).Name(), ".")
			name = nameParts[len(nameParts)-1]
		} else {
			name = typ.Elem().Name()
		}
	}
	_, exists := ep.extensions[name]
	if exists {
		return false
	}
	ep.extensions[name] = extension
	return true
}

func (ep *extensionPoint) unregister(name string) bool {
	ep.Lock()
	defer ep.Unlock()
	_, exists := ep.extensions[name]
	if !exists {
		return false
	}
	delete(ep.extensions, name)
	return true
}

// Endpoint

var Endpoints = &endpointExt{
	newExtensionPoint(new(Endpoint)),
}

type endpointExt struct {
	*extensionPoint
}

func (ep *endpointExt) Unregister(name string) bool {
	return ep.unregister(name)
}

func (ep *endpointExt) Register(extension Endpoint, name string) bool {
	return ep.register(extension, name)
}

func (ep *endpointExt) Lookup(name string) Endpoint {
	ext := ep.lookup(name)
	if ext == nil {
		return nil
	}
	return ext.(Endpoint)
}

func (ep *endpointExt) Select(names []string) []Endpoint {
	var selected []Endpoint
	for _, name := range names {
		selected = append(selected, ep.Lookup(name))
	}
	return selected
}

func (ep *endpointExt) All() map[string]Endpoint {
	all := make(map[string]Endpoint)
	for k, v := range ep.all() {
		all[k] = v.(Endpoint)
	}
	return all
}

func (ep *endpointExt) Names() []string {
	var names []string
	for k := range ep.all() {
		names = append(names, k)
	}
	return names
}
