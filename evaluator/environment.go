package evaluator

import "fmt"

// Environment represents a scope in the Glace runtime.
// Each environment has a reference to its parent (enclosing) scope,
// forming a chain for lexical scoping.
type Environment struct {
	store   map[string]binding
	parent  *Environment
}

// binding holds a value and its mutability flag.
type binding struct {
	value   Value
	mutable bool
}

// NewEnvironment creates a new root environment with no parent.
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]binding),
	}
}

// NewEnclosedEnvironment creates a child environment with the given parent.
// Used for function scopes, block scopes, and loop scopes.
func NewEnclosedEnvironment(parent *Environment) *Environment {
	return &Environment{
		store:  make(map[string]binding),
		parent: parent,
	}
}

// Define creates a new binding in the CURRENT scope.
// Returns an error if the variable is already defined in this scope.
func (e *Environment) Define(name string, val Value, mutable bool) error {
	if _, exists := e.store[name]; exists {
		return fmt.Errorf("variable '%s' is already defined in this scope", name)
	}
	e.store[name] = binding{value: val, mutable: mutable}
	return nil
}

// Get looks up a variable by name, walking up the scope chain.
// Returns the value and true if found, or nil and false if not.
func (e *Environment) Get(name string) (Value, bool) {
	b, ok := e.store[name]
	if ok {
		return b.value, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set updates an existing variable's value.
// Walks the scope chain to find the binding.
// Returns an error if the variable is not found or is immutable.
func (e *Environment) Set(name string, val Value) error {
	b, ok := e.store[name]
	if ok {
		if !b.mutable {
			return fmt.Errorf("cannot assign to immutable variable '%s'", name)
		}
		e.store[name] = binding{value: val, mutable: true}
		return nil
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return fmt.Errorf("undefined variable '%s'", name)
}

// IsMutable returns whether a variable is declared as mutable.
func (e *Environment) IsMutable(name string) bool {
	b, ok := e.store[name]
	if ok {
		return b.mutable
	}
	if e.parent != nil {
		return e.parent.IsMutable(name)
	}
	return false
}
