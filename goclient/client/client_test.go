/*
 * Copyright 2016 Dgraph Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"testing"

	"github.com/dgraph-io/dgraph/query/graph"
	"github.com/dgraph-io/dgraph/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckNQuad(t *testing.T) {
	if err := checkNQuad(graph.NQuad{
		Predicate:   "name",
		ObjectValue: types.Str("Alice"),
	}); err == nil {
		t.Fatal(err)
	}
	if err := checkNQuad(graph.NQuad{
		Subject:     "alice",
		ObjectValue: types.Str("Alice"),
	}); err == nil {
		t.Fatal(err)
	}
	if err := checkNQuad(graph.NQuad{
		Subject:   "alice",
		Predicate: "name",
	}); err == nil {
		t.Fatal(err)
	}
	if err := checkNQuad(graph.NQuad{
		Subject:     "alice",
		Predicate:   "name",
		ObjectValue: types.Str("Alice"),
		ObjectId:    "id",
	}); err == nil {
		t.Fatal(err)
	}
}

func TestSetMutation(t *testing.T) {
	req := Req{}

	if err := req.AddMutation(graph.NQuad{
		Subject:     "alice",
		Predicate:   "name",
		ObjectValue: types.Str("Alice"),
	}, SET); err != nil {
		t.Fatal(err)
	}

	if err := req.AddMutation(graph.NQuad{
		Subject:     "alice",
		Predicate:   "falls.in",
		ObjectValue: types.Str("rabbithole"),
	}, SET); err != nil {
		t.Fatal(err)
	}

	if err := req.AddMutation(graph.NQuad{
		Subject:     "alice",
		Predicate:   "falls.in",
		ObjectValue: types.Str("rabbithole"),
	}, DEL); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(req.gr.Mutation.Set), 2, "Set should have 2 entries")
	assert.Equal(t, len(req.gr.Mutation.Del), 1, "Del should have 1 entry")
}