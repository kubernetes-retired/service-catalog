/*
Copyright 2018 The Kubernetes Authors.

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

package requirements

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestEverything(t *testing.T) {
	if !Everything().Matches(Set{"x": "y"}) {
		t.Errorf("Nil selector didn't match")
	}
	if !Everything().Empty() {
		t.Errorf("Everything was not empty")
	}
}

func TestRequirementConstructor(t *testing.T) {
	requirementConstructorTests := []struct {
		Key     Term
		Op      selection.Operator
		Vals    sets.String
		Success bool
	}{
		{"x", selection.In, nil, false},
		{"x", selection.NotIn, sets.NewString(), false},
		{"x", selection.In, sets.NewString("foo"), true},
		{"x", selection.NotIn, sets.NewString("foo"), true},
		{"x", selection.Exists, nil, true},
		{"x", selection.DoesNotExist, nil, true},
		{"1foo", selection.In, sets.NewString("bar"), true},
		{"1234", selection.In, sets.NewString("bar"), true},
	}
	for _, rc := range requirementConstructorTests {
		if _, err := NewRequirement(rc.Key, rc.Op, rc.Vals.List()); err == nil && !rc.Success {
			t.Errorf("expected error with key:%#v op:%v vals:%v, got no error", rc.Key, rc.Op, rc.Vals)
		} else if err != nil && rc.Success {
			t.Errorf("expected no error with key:%#v op:%v vals:%v, got:%v", rc.Key, rc.Op, rc.Vals, err)
		}
	}
}

func TestToString(t *testing.T) {
	var req Requirement
	toStringTests := []struct {
		In    *internalPredicate
		Out   string
		Valid bool
	}{

		{&internalPredicate{
			getRequirement("x", selection.In, sets.NewString("abc", "def"), t),
			getRequirement("y", selection.NotIn, sets.NewString("jkl"), t),
			getRequirement("z", selection.Exists, nil, t)},
			"x in (abc,def),y notin (jkl),z", true},
		{&internalPredicate{
			getRequirement("x", selection.NotIn, sets.NewString("abc", "def"), t),
			getRequirement("y", selection.NotEquals, sets.NewString("jkl"), t),
			getRequirement("z", selection.DoesNotExist, nil, t)},
			"x notin (abc,def),y!=jkl,!z", true},
		{&internalPredicate{
			getRequirement("x", selection.In, sets.NewString("abc", "def"), t),
			req}, // adding empty req for the trailing ','
			"x in (abc,def),", false},
		{&internalPredicate{
			getRequirement("x", selection.NotIn, sets.NewString("abc"), t),
			getRequirement("y", selection.In, sets.NewString("jkl", "mno"), t),
			getRequirement("z", selection.NotIn, sets.NewString(""), t)},
			"x notin (abc),y in (jkl,mno),z notin ()", true},
		{&internalPredicate{
			getRequirement("x", selection.Equals, sets.NewString("abc"), t),
			getRequirement("y", selection.DoubleEquals, sets.NewString("jkl"), t),
			getRequirement("z", selection.NotEquals, sets.NewString("a"), t),
			getRequirement("z", selection.Exists, nil, t)},
			"x=abc,y==jkl,z!=a,z", true},
	}
	for _, ts := range toStringTests {
		if out := ts.In.String(); out == "" && ts.Valid {
			t.Errorf("%#v.String() => '%v' expected no error", ts.In, out)
		} else if out != ts.Out {
			t.Errorf("%#v.String() => '%v' want '%v'", ts.In, out, ts.Out)
		}
	}
}

func TestRequirementPredicateMatching(t *testing.T) {
	var req Requirement
	labelPredicateMatchingTests := []struct {
		Set   Set
		Sel   Predicate
		Match bool
	}{
		{Set{"x": "foo", "y": "baz"}, &internalPredicate{
			req,
		}, false},
		{Set{"x": "foo", "y": "baz"}, &internalPredicate{
			getRequirement("x", selection.In, sets.NewString("foo"), t),
			getRequirement("y", selection.NotIn, sets.NewString("alpha"), t),
		}, true},
		{Set{"x": "foo", "y": "baz"}, &internalPredicate{
			getRequirement("x", selection.In, sets.NewString("foo"), t),
			getRequirement("y", selection.In, sets.NewString("alpha"), t),
		}, false},
		{Set{"y": ""}, &internalPredicate{
			getRequirement("x", selection.NotIn, sets.NewString(""), t),
			getRequirement("y", selection.Exists, nil, t),
		}, true},
		{Set{"y": ""}, &internalPredicate{
			getRequirement("x", selection.DoesNotExist, nil, t),
			getRequirement("y", selection.Exists, nil, t),
		}, true},
		{Set{"y": ""}, &internalPredicate{
			getRequirement("x", selection.NotIn, sets.NewString(""), t),
			getRequirement("y", selection.DoesNotExist, nil, t),
		}, false},
		{Set{"y": "baz"}, &internalPredicate{
			getRequirement("x", selection.In, sets.NewString(""), t),
		}, false},
	}
	for _, lsm := range labelPredicateMatchingTests {
		if match := lsm.Sel.Matches(lsm.Set); match != lsm.Match {
			t.Errorf("%+v.Matches(%#v) => %v, want %v", lsm.Sel, lsm.Set, match, lsm.Match)
		}
	}
}

func getRequirement(key Term, op selection.Operator, vals sets.String, t *testing.T) Requirement {
	req, err := NewRequirement(key, op, vals.List())
	if err != nil {
		t.Errorf("NewRequirement(%v, %v, %v) resulted in error:%v", key, op, vals, err)
		return Requirement{}
	}
	return *req
}

func TestAdd(t *testing.T) {
	testCases := []struct {
		name         string
		sel          Predicate
		key          Term
		operator     selection.Operator
		values       []string
		refPredicate Predicate
	}{
		{
			"keyInOperator",
			internalPredicate{},
			"key",
			selection.In,
			[]string{"value"},
			internalPredicate{Requirement{"key", selection.In, []string{"value"}}},
		},
		{
			"keyEqualsOperator",
			internalPredicate{Requirement{"key", selection.In, []string{"value"}}},
			"key2",
			selection.Equals,
			[]string{"value2"},
			internalPredicate{
				Requirement{"key", selection.In, []string{"value"}},
				Requirement{"key2", selection.Equals, []string{"value2"}},
			},
		},
	}
	for _, ts := range testCases {
		req, err := NewRequirement(ts.key, ts.operator, ts.values)
		if err != nil {
			t.Errorf("%s - Unable to create labels.Requirement", ts.name)
		}
		ts.sel = ts.sel.Add(*req)
		if !reflect.DeepEqual(ts.sel, ts.refPredicate) {
			t.Errorf("%s - Expected %v found %v", ts.name, ts.refPredicate, ts.sel)
		}
	}
}
