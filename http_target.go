// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trunks

import (
	"path/filepath"
	"sync"

	libhttp "github.com/shuLhan/share/lib/http"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

//
// HttpRunHandler define the function type that will be called when client
// send request to run the HTTP target.
//
type HttpRunHandler func(rr *RunRequest) ([]byte, error)

//
// HttpAttackHandler define the function type that will be called when client
// send request to attack HTTP target.
//
type HttpAttackHandler func(rr *RunRequest) vegeta.Targeter

// HttpPreAttackHandler define the function type that will be called before
// the actual Attack being called.
type HttpPreAttackHandler func(rr *RunRequest)

type HttpTarget struct {
	// ID of target, optional.
	// If its empty, it will generated using value from Path.
	ID string

	// Name of target, optional, default to Path.
	Name string

	Method      libhttp.RequestMethod
	Path        string
	RequestType libhttp.RequestType
	Headers     KeyValue
	Params      KeyValue

	Run          HttpRunHandler       `json:"-"`
	Attack       HttpAttackHandler    `json:"-"`
	PreAttack    HttpPreAttackHandler `json:"-"`
	AttackLocker sync.Mutex           `json:"-"` // Use this inside the Attack to lock resource.
	Status       string

	// Results contains list of load testing output.
	Results []*AttackResult

	// AllowAttack if its true the "Attack" button will be showed on user
	// interface to allow client to run load testing on this HttpTarget.
	AllowAttack bool
}

func (ht *HttpTarget) init() {
	if len(ht.ID) == 0 {
		ht.ID = generateID(ht.Path)
	}
	if len(ht.Name) == 0 {
		ht.Name = ht.Path
	}
	if ht.Headers == nil {
		ht.Headers = KeyValue{}
	}
	if ht.Params == nil {
		ht.Params = KeyValue{}
	}
}

func (ht *HttpTarget) deleteResult(result *AttackResult) {
	var x int
	for ; x < len(ht.Results); x++ {
		if ht.Results[x].Name == result.Name {
			break
		}
	}

	if x == len(ht.Results) {
		// Not found.
		return
	}

	copy(ht.Results[x:], ht.Results[x+1:])
	ht.Results[len(ht.Results)-1] = nil
	ht.Results = ht.Results[:len(ht.Results)-1]
}

func (ht *HttpTarget) addResult(dir, name string) (err error) {
	ar := &AttackResult{
		HttpTargetID: ht.ID,
		Name:         name,
		fullpath:     filepath.Join(dir, name),
	}

	ht.Results = append(ht.Results, ar)

	return nil
}

func (ht *HttpTarget) getResultByName(name string) (result *AttackResult) {
	for _, result = range ht.Results {
		if result.Name == name {
			return result
		}
	}
	return nil
}
