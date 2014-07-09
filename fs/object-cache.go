// Licensed to Elasticsearch under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package fs

import (
	"fmt"
	"github.com/elasticsearch/kriterium/panics"
	"log"
	"strconv"
	"time"
)

type InfoAge time.Duration

func (t *InfoAge) String() string {
	return fmt.Sprintf("%d", *t)
}

func (t *InfoAge) Set(vrep string) error {
	v, e := strconv.ParseInt(vrep, 10, 64)
	if e != nil {
		return e
	}
	var tt time.Duration = time.Duration(v) * time.Millisecond
	*t = InfoAge(tt)
	return nil
}

// fs.Object cache
type ObjectCache struct {
	options struct {
		maxSize uint16
		maxAge  InfoAge
	}
	Cache  map[string]Object
	gc     GcFunc
	gcArgs []interface{}
}

var GCAlgorithm = struct{ byAge, bySize GcFunc }{
	byAge:  AgeBasedGcFunc,
	bySize: SizeBasedGcFunc,
}

func NewFixedSizeObjectCache(maxSize uint16) *ObjectCache {
	oc := newObjectCache()
	oc.options.maxSize = maxSize
	oc.gc = SizeBasedGcFunc
	oc.gcArgs = []interface{}{oc.options.maxSize}
	return oc
}
func NewTimeWindowObjectCache(maxAge InfoAge) *ObjectCache {
	oc := newObjectCache()
	oc.options.maxAge = maxAge
	oc.gc = AgeBasedGcFunc
	oc.gcArgs = []interface{}{oc.options.maxAge}
	return oc
}
func newObjectCache() *ObjectCache {
	oc := new(ObjectCache)
	oc.Cache = make(map[string]Object)
	return oc
}
func (oc *ObjectCache) MarkDeleted(id string) bool {
	obj, found := oc.Cache[id]
	if !found {
		return false
	}
	obj.SetFlags(1)
	return true
}
func (oc *ObjectCache) IsDeleted(id string) (bool, error) {
	obj, found := oc.Cache[id]
	if !found {
		return false, ERR.OBJECT_NOT_FOUND(id)
	}

	return IsDeleted0(obj.Flags()), nil
}

func IsDeleted0(flag uint8) bool {
	return flag == uint8(1)
}

func (oc *ObjectCache) Gc() {
	n := oc.gc(oc.Cache, oc.gcArgs...)
	if n > 0 {
		log.Printf("DEBUG: GC: %d items removed - object-cnt: %d", n, len(oc.Cache))
	}
}

// REVU: TODO: sort these by age descending first
func AgeBasedGcFunc(cache map[string]Object, args ...interface{}) int {
	panics.OnFalse(len(args) == 1, "BUG", "AgeBasedGcFunc", "args:", args)
	limit, ok := args[0].(InfoAge)
	panics.OnFalse(ok, "BUG", "AgeBasedGcFunc", "limit:", args[0])
	n := 0
	for id, obj := range cache {
		if !IsDeleted0(obj.Flags()) {
			continue
		}
		if obj.Age() > time.Duration(limit) {
			delete(cache, id)
			n++
		}
	}
	return n
}

func SizeBasedGcFunc(cache map[string]Object, args ...interface{}) int {
	panics.OnFalse(len(args) == 1, "BUG", "SizeBasedGcFunc", "args:", args)
	limit, ok := args[0].(uint16)
	panics.OnFalse(ok, "BUG", "SizeBasedGcFunc", "limit:", args[0])
	n := 0
	if len(cache) <= int(limit) {
		return 0
	}
	for id, obj := range cache {
		if !IsDeleted0(obj.Flags()) {
			continue
		}
		delete(cache, id)
		n++
	}
	return n
}

type GcFunc func(cache map[string]Object, args ...interface{}) int
