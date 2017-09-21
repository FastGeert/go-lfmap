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

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_basic(t *testing.T) {
	myMap := NewLFmap()

	assert.Equal(t, false, myMap.Exists("key1"))

	myMap.Set("key1", "value1")

	assert.Equal(t, true, myMap.Exists("key1"))

	if tmpValueAsInterface, exists := myMap.Get("key1"); exists {
		if valueAsString, ok := tmpValueAsInterface.(string); ok {
			assert.Equal(t, valueAsString, "value1")
		} else {
			t.Fail()
		}
	} else {
		t.Fail()
	}

	assert.Equal(t, true, myMap.Exists("key1"))

	myMap.Set("key1", "value2")

	if tmpValueAsInterface, exists := myMap.Get("key1"); exists {
		if valueAsString, ok := tmpValueAsInterface.(string); ok {
			assert.Equal(t, valueAsString, "value2")
		} else {
			t.Fail()
		}
	} else {
		t.Fail()
	}

	myMap.Remove("key1")

	assert.Equal(t, false, myMap.Exists("key1"))

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Test_concurrent(t *testing.T) {
	myMap := NewLFmap()

	type entry struct {
		key   string
		value string
	}

	channelDone := make(chan bool)

	leftChannel := make(chan *entry)
	rightChannel := make(chan *entry)

	for i := 0; i < 1000; i++ {
		go func(index int) {
			key := fmt.Sprintf("key%d", index)
			value := RandStringBytes(32)
			assert.Equal(t, false, myMap.Exists(key), "Key %s already exists!", key)

			myMap.Set(key, value)

			e := new(entry)
			e.key = key
			e.value = value

			if index%2 == 0 {
				leftChannel <- e
				e = <-rightChannel
			} else {
				leftE := <-leftChannel
				rightChannel <- e
				e = leftE
			}

			assert.Equal(t, true, myMap.Exists(e.key))

			if tmpValueAsInterface, exists := myMap.Get(e.key); exists {
				if valueAsString, ok := tmpValueAsInterface.(string); ok {
					assert.Equal(t, valueAsString, e.value)
				} else {
					t.Error()
				}
			} else {
				t.Fail()
			}

			assert.Equal(t, true, myMap.Exists(key))

			if tmpValueAsInterface, exists := myMap.Get(key); exists {
				if valueAsString, ok := tmpValueAsInterface.(string); ok {
					assert.Equal(t, valueAsString, value)
				} else {
					t.Fail()
				}
			} else {
				t.Fail()
			}

			assert.Equal(t, true, myMap.Exists(key))

			channelDone <- true
		}(i)
	}

	// Wait for all the tests to complete
	for i := 0; i < 1000; i++ {
		<-channelDone
	}
}

func Test_speed(t *testing.T) {

	data := [1000000]string{}
	for i := 0; i < 1000; i++ {
		data[i] = fmt.Sprintf("test: %d", i)
	}

	start := time.Now()
	nm := make(map[string]string)
	for i := 0; i < 1000000; i++ {
		d := data[i]
		nm[d] = d
	}
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stdout, "1000000 sets to normal map take %s seconds\n", elapsed)

	doneChannel := make(chan bool, 10)
	start = time.Now()
	lfm := NewLFmap()
	for i := 0; i < 100; i++ {
		go func(multiplier int) {
			for i := 0; i < 10000; i++ {
				index := 10*multiplier + i
				d := data[index]
				lfm.Set(d, d)
				doneChannel <- true
			}
		}(i)
	}
	for i := 0; i < 1000000; i++ {
		<-doneChannel
	}
	elapsed = time.Since(start)
	fmt.Fprintf(os.Stdout, "1000000 sets to lock free map take %s seconds\n", elapsed)

	start = time.Now()
	mutex := new(sync.RWMutex)
	for i := 0; i < 100; i++ {
		go func(multiplier int) {
			for i := 0; i < 10000; i++ {
				index := 10*multiplier + i
				d := data[index]
				mutex.Lock()
				nm[d] = d
				mutex.Unlock()
				doneChannel <- true
			}
		}(i)
	}
	for i := 0; i < 1000000; i++ {
		<-doneChannel
	}
	elapsed = time.Since(start)
	fmt.Fprintf(os.Stdout, "1000000 sets to normal protected with mutex map take %s seconds\n", elapsed)

}
