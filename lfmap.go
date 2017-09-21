/*
   Copyright 2017 GIG Technology NV

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package lfmap

const (
	get    = 1
	set    = 2
	remove = 3
	exists = 4
	stop   = 5
)

type entry struct {
	key   string
	value interface{}
}

type command struct {
	operation int
	data      *entry
}

type getResult struct {
	value  interface{}
	exists bool
}

// LFmap provides a lock free threadsafe hashmap
type LFmap struct {
	commandChannel chan *command
	getChannel     chan *getResult
	existsChannel  chan bool
	running        bool
}

// NewLFmap constructs an LFmap hashmap
func NewLFmap() *LFmap {
	m := new(LFmap)
	m.commandChannel = make(chan *command)
	m.getChannel = make(chan *getResult)
	m.existsChannel = make(chan bool)

	go func() {
		im := make(map[string]*entry)
		for {
			c := <-m.commandChannel
			switch c.operation {
			case get:
				e, exists := im[c.data.key]

				gr := new(getResult)
				if exists {
					gr.value = e.value
				}
				gr.exists = exists
				m.getChannel <- gr
			case set:
				im[c.data.key] = c.data
			case remove:
				delete(im, c.data.key)
			case exists:
				_, exists := im[c.data.key]
				m.existsChannel <- exists
			case stop:
				break
			}
		}
	}()

	m.running = true

	return m
}

// Get retrieves a value with key from the hashmap
func (lfmap *LFmap) Get(key string) (interface{}, bool) {
	if !lfmap.running {
		panic("Cannot interact with a non running LFmap!")
	}
	c := new(command)
	c.operation = get
	d := new(entry)
	d.key = key
	c.data = d

	lfmap.commandChannel <- c
	gr := <-lfmap.getChannel
	return gr.value, gr.exists
}

// Set sets the value for key in the hashmap
func (lfmap *LFmap) Set(key string, value interface{}) {
	if !lfmap.running {
		panic("Cannot interact with a non running LFmap!")
	}
	c := new(command)
	d := new(entry)
	c.operation = set
	d.key = key
	d.value = value
	c.data = d

	lfmap.commandChannel <- c
}

// Remove removes the value for key from the hashmap
func (lfmap *LFmap) Remove(key string) {
	if !lfmap.running {
		panic("Cannot interact with a non running LFmap!")
	}
	c := new(command)
	c.operation = remove
	d := new(entry)
	d.key = key
	c.data = d

	lfmap.commandChannel <- c
}

// Exists checks for the existence of a certain key in the hashmap
func (lfmap *LFmap) Exists(key string) bool {
	if !lfmap.running {
		panic("Cannot interact with a non running LFmap!")
	}
	c := new(command)
	c.operation = exists
	d := new(entry)
	d.key = key
	c.data = d

	lfmap.commandChannel <- c
	return <-lfmap.existsChannel
}

// Stop stops the hashmap, and removes all keys from memory. After calling Stop() the hashmap becomes unusable and will panic on every method call except for IsRunning
func (lfmap *LFmap) Stop() {
	if !lfmap.running {
		panic("Cannot interact with a non running LFmap!")
	}
	c := new(command)
	c.operation = stop

	lfmap.commandChannel <- c
	lfmap.running = false
}

// IsRunning checks if the hashmap is still running
func (lfmap *LFmap) IsRunning() bool {
	return lfmap.running
}
