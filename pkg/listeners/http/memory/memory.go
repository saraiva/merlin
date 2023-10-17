// Merlin is a post-exploitation command and control framework.
// This file is part of Merlin.
// Copyright (C) 2023  Russel Van Tuyl

// Merlin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.

// Merlin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Merlin.  If not, see <http://www.gnu.org/licenses/>.

// Package memory is an in-memory database used to store and retrieve HTTP listeners
package memory

import (
	// Standard
	"fmt"
	"sync"

	// 3rd Party
	"github.com/google/uuid"

	// Merlin
	"github.com/Ne0nd0g/merlin/pkg/listeners/http"
)

// Repository is a structure that implements the Repository interface
type Repository struct {
	listeners map[uuid.UUID]http.Listener
	sync.Mutex
}

// listenerMap is the in-memory structure that holds a map of created and stored HTTP listeners
var listenerMap = make(map[uuid.UUID]http.Listener)

// NewRepository is a factory to create and return a repository object to store and manage listeners
func NewRepository() *Repository {
	return &Repository{
		listeners: listenerMap,
		Mutex:     sync.Mutex{},
	}
}

// Add stores the passed in HTTP listener
func (r *Repository) Add(listener http.Listener) error {
	// Make sure the map exists and create it if not
	if r.listeners == nil {
		r.Lock()
		r.listeners = make(map[uuid.UUID]http.Listener)
		r.Unlock()
	}
	// Make sure the listener isn't already in the map
	if _, ok := r.listeners[listener.ID()]; ok {
		return fmt.Errorf("a listener with an ID of %s already exists", listener.ID())
	}
	// Add
	r.Lock()
	r.listeners[listener.ID()] = listener
	r.Unlock()
	return nil
}

// Exists determines if the HTTP listener has already been instantiated
func (r *Repository) Exists(name string) bool {
	for _, l := range r.listeners {
		if name == l.Name() {
			return true
		}
	}
	return false
}

// List returns a list of Listeners that exist and is used for command line tab completion
func (r *Repository) List() func(string) []string {
	return func(line string) []string {
		var l []string
		for _, listener := range r.listeners {
			l = append(l, listener.Name())
		}
		return l
	}
}

// Listeners returns a list of all stored Listener objects to be consumed by a client application
func (r *Repository) Listeners() []http.Listener {
	var found []http.Listener
	for _, l := range r.listeners {
		found = append(found, l)
	}
	return found
}

// ListenerByID finds and returns the listener object by its ID (UUIDv4)
func (r *Repository) ListenerByID(id uuid.UUID) (http.Listener, error) {
	l, exists := r.listeners[id]
	if !exists {
		return http.Listener{}, fmt.Errorf(fmt.Sprintf("a listener with an ID of %s does not exist", id))
	}
	return l, nil
}

// ListenerByName finds and returns  the listener object by its name (string)
func (r *Repository) ListenerByName(name string) (http.Listener, error) {
	if !r.Exists(name) {
		return http.Listener{}, fmt.Errorf("%s listener does not exist", name)
	}

	var listener http.Listener
	for _, l := range r.listeners {
		if name == l.Name() {
			listener = l
			break
		}
	}
	return listener, nil
}

// RemoveByID deletes a listener from the global list of Listeners by the input UUID
func (r *Repository) RemoveByID(id uuid.UUID) error {
	if l, ok := r.listeners[id]; ok {
		/*
			err := l.Stop()
			if err != nil {
				return err
			}
		*/
		delete(r.listeners, l.ID())
		return nil
	}
	return fmt.Errorf("could not remove listener: %s because it does not exist", id)
}

// SetOption updates the listener's configurable options value passed in
func (r *Repository) SetOption(id uuid.UUID, option, value string) error {
	listener, err := r.ListenerByID(id)
	if err != nil {
		return fmt.Errorf("pkg/listeners/http/memory.SetOption(): %s", err)
	}
	r.Lock()
	defer r.Unlock()
	err = listener.SetOption(option, value)
	if err != nil {
		return fmt.Errorf("pkg/listeners/http/memory.SetOption(): %s", err)
	}
	r.listeners[listener.ID()] = listener
	return nil
}
