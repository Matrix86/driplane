package plugin

import (
	"fmt"

	"github.com/dop251/goja"
)

func (p *Plugin) compile() error {
	// Create a new goja runtime
	p.vm = goja.New()

	// Set predefined objects and functions
	for name, def := range Defines {
		if err := p.vm.Set(name, def); err != nil {
			return fmt.Errorf("error setting predefined '%s': %s", name, err)
		}
	}

	// Run the plugin code
	_, err := p.vm.RunString(p.Code)
	if err != nil {
		return fmt.Errorf("error compiling plugin: %s", err)
	}

	// Extract callbacks (functions) and objects from the global object
	globalObj := p.vm.GlobalObject()
	for _, key := range globalObj.Keys() {
		val := globalObj.Get(key)
		
		if val == nil || val == goja.Undefined() || val == goja.Null() {
			continue
		}

		// Check if it's a function
		if fn, ok := goja.AssertFunction(val); ok {
			p.callbacks[key] = fn
		} else {
			// It's an object or primitive
			p.objects[key] = val
		}
	}

	return nil
}
