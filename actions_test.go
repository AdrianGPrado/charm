// Copyright 2011-2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package charm

import (
	"bytes"
	"encoding/json"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type ActionsSuite struct{}

var _ = gc.Suite(&ActionsSuite{})

func (s *ActionsSuite) TestNewActions(c *gc.C) {
	emptyAction := NewActions()
	c.Assert(emptyAction, jc.DeepEquals, &Actions{})
}

func (s *ActionsSuite) TestValidateOk(c *gc.C) {
	for i, test := range []struct {
		description      string
		actionSpec       *ActionSpec
		objectToValidate map[string]interface{}
	}{{
		description: "Validation of an empty object is ok.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"}}}},
		objectToValidate: nil,
	}, {
		description: "Validation of one required value.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"}},
				"required": []interface{}{"outfile"}}},
		objectToValidate: map[string]interface{}{
			"outfile": "out-2014-06-12.bz2",
		},
	}, {
		description: "Validation of one required and one optional value.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"},
					"quality": map[string]interface{}{
						"description": "Compression quality",
						"type":        "integer",
						"minimum":     0,
						"maximum":     9}},
				"required": []interface{}{"outfile"}}},
		objectToValidate: map[string]interface{}{
			"outfile": "out-2014-06-12.bz2",
		},
	}, {
		description: "Validation of an optional, range limited value.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"},
					"quality": map[string]interface{}{
						"description": "Compression quality",
						"type":        "integer",
						"minimum":     0,
						"maximum":     9}},
				"required": []interface{}{"outfile"}}},
		objectToValidate: map[string]interface{}{
			"outfile": "out-2014-06-12.bz2",
			"quality": 5,
		},
	}} {
		c.Logf("test %d: %s", i, test.description)
		err := test.actionSpec.ValidateParams(test.objectToValidate)
		c.Assert(err, jc.ErrorIsNil)
	}
}

func (s *ActionsSuite) TestValidateFail(c *gc.C) {
	var validActionTests = []struct {
		description   string
		actionSpec    *ActionSpec
		badActionJson string
		expectedError string
	}{{
		description: "Validation of one required value.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"}},
				"required": []interface{}{"outfile"}}},
		badActionJson: `{"outfile": 5}`,
		expectedError: "validation failed: (root).outfile : must be of type string, given 5",
	}, {
		description: "Restrict to only one property",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"}},
				"required":             []interface{}{"outfile"},
				"additionalProperties": false}},
		badActionJson: `{"outfile": "foo.bz", "bar": "foo"}`,
		expectedError: "validation failed: (root) : additional property \"bar\" is not allowed, given {\"bar\":\"foo\",\"outfile\":\"foo.bz\"}",
	}, {
		description: "Validation of one required and one optional value.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"},
					"quality": map[string]interface{}{
						"description": "Compression quality",
						"type":        "integer",
						"minimum":     0,
						"maximum":     9}},
				"required": []interface{}{"outfile"}}},
		badActionJson: `{"quality": 5}`,
		expectedError: "validation failed: (root) : \"outfile\" property is missing and required, given {\"quality\":5}",
	}, {
		description: "Validation of an optional, range limited value.",
		actionSpec: &ActionSpec{
			Description: "Take a snapshot of the database.",
			Params: map[string]interface{}{
				"title":       "snapshot",
				"description": "Take a snapshot of the database.",
				"type":        "object",
				"properties": map[string]interface{}{
					"outfile": map[string]interface{}{
						"description": "The file to write out to.",
						"type":        "string"},
					"quality": map[string]interface{}{
						"description": "Compression quality",
						"type":        "integer",
						"minimum":     0,
						"maximum":     9}},
				"required": []interface{}{"outfile"}}},
		badActionJson: `
{ "outfile": "out-2014-06-12.bz2", "quality": "two" }`,
		expectedError: "validation failed: (root).quality : must be of type integer, given \"two\"",
	}}

	for i, test := range validActionTests {
		c.Logf("test %d: %s", i, test.description)
		var params map[string]interface{}
		jsonBytes := []byte(test.badActionJson)
		err := json.Unmarshal(jsonBytes, &params)
		c.Assert(err, gc.IsNil)
		err = test.actionSpec.ValidateParams(params)
		c.Assert(err.Error(), gc.Equals, test.expectedError)
	}
}

func (s *ActionsSuite) TestCleanseOk(c *gc.C) {

	var goodInterfaceTests = []struct {
		description         string
		acceptableInterface map[string]interface{}
		expectedInterface   map[string]interface{}
	}{{
		description: "An interface requiring no changes.",
		acceptableInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
		expectedInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
	}, {
		description: "Substitute a single inner map[i]i.",
		acceptableInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[interface{}]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
		expectedInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{
				"foo1": "val1",
				"foo2": "val2"}},
	}, {
		description: "Substitute nested inner map[i]i.",
		acceptableInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": "val2a",
			"key3a": map[interface{}]interface{}{
				"key1b": "val1b",
				"key2b": map[interface{}]interface{}{
					"key1c": "val1c"}}},
		expectedInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": "val2a",
			"key3a": map[string]interface{}{
				"key1b": "val1b",
				"key2b": map[string]interface{}{
					"key1c": "val1c"}}},
	}, {
		description: "Substitute nested map[i]i within []i.",
		acceptableInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": []interface{}{5, "foo", map[string]interface{}{
				"key1b": "val1b",
				"key2b": map[interface{}]interface{}{
					"key1c": "val1c"}}}},
		expectedInterface: map[string]interface{}{
			"key1a": "val1a",
			"key2a": []interface{}{5, "foo", map[string]interface{}{
				"key1b": "val1b",
				"key2b": map[string]interface{}{
					"key1c": "val1c"}}}},
	}}

	for i, test := range goodInterfaceTests {
		c.Logf("test %d: %s", i, test.description)
		cleansedInterfaceMap, err := cleanse(test.acceptableInterface)
		c.Assert(err, gc.IsNil)
		c.Assert(cleansedInterfaceMap, jc.DeepEquals, test.expectedInterface)
	}
}

func (s *ActionsSuite) TestCleanseFail(c *gc.C) {

	var badInterfaceTests = []struct {
		description   string
		failInterface map[string]interface{}
		expectedError string
	}{{
		description: "An inner map[interface{}]interface{} with an int key.",
		failInterface: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[interface{}]interface{}{
				"foo1": "val1",
				5:      "val2"}},
		expectedError: "map keyed with non-string value",
	}, {
		description: "An inner []interface{} containing a map[i]i with an int key.",
		failInterface: map[string]interface{}{
			"key1a": "val1b",
			"key2a": "val2b",
			"key3a": []interface{}{"foo1", 5, map[interface{}]interface{}{
				"key1b": "val1b",
				"key2b": map[interface{}]interface{}{
					"key1c": "val1c",
					5:       "val2c"}}}},
		expectedError: "map keyed with non-string value",
	}}

	for i, test := range badInterfaceTests {
		c.Logf("test %d: %s", i, test.description)
		_, err := cleanse(test.failInterface)
		c.Assert(err, gc.NotNil)
		c.Assert(err.Error(), gc.Equals, test.expectedError)
	}
}

func (s *ActionsSuite) TestReadGoodActionsYaml(c *gc.C) {
	var goodActionsYamlTests = []struct {
		description     string
		yaml            string
		expectedActions *Actions
	}{{
		description: "A simple snapshot actions YAML with one parameter.",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params:
      outfile:
         description: "The file to write out to."
         type: string
   required: ["outfile"]
`,
		expectedActions: &Actions{map[string]ActionSpec{
			"snapshot": ActionSpec{
				Description: "Take a snapshot of the database.",
				Params: map[string]interface{}{
					"title":       "snapshot",
					"description": "Take a snapshot of the database.",
					"type":        "object",
					"properties": map[string]interface{}{
						"outfile": map[string]interface{}{
							"description": "The file to write out to.",
							"type":        "string"}},
					"required": []interface{}{"outfile"}}}}},
	}, {
		description: "An empty Actions definition.",
		yaml:        "",
		expectedActions: &Actions{
			ActionSpecs: map[string]ActionSpec{},
		},
	}, {
		description: "A more complex schema with hyphenated names and multiple parameters.",
		yaml: `
snapshot:
   description: "Take a snapshot of the database."
   params:
      outfile:
         description: "The file to write out to."
         type: "string"
      compression-quality:
         description: "The compression quality."
         type: "integer"
         minimum: 0
         maximum: 9
         exclusiveMaximum: false
remote-sync:
   description: "Sync a file to a remote host."
   params:
      file:
         description: "The file to send out."
         type: "string"
         format: "uri"
      remote-uri:
         description: "The host to sync to."
         type: "string"
         format: "uri"
      util:
         description: "The util to perform the sync (rsync or scp.)"
         type: "string"
         enum: ["rsync", "scp"]
   required: ["file", "remote-uri"]
`,
		expectedActions: &Actions{map[string]ActionSpec{
			"snapshot": ActionSpec{
				Description: "Take a snapshot of the database.",
				Params: map[string]interface{}{
					"title":       "snapshot",
					"description": "Take a snapshot of the database.",
					"type":        "object",
					"properties": map[string]interface{}{
						"outfile": map[string]interface{}{
							"description": "The file to write out to.",
							"type":        "string"},
						"compression-quality": map[string]interface{}{
							"description":      "The compression quality.",
							"type":             "integer",
							"minimum":          0,
							"maximum":          9,
							"exclusiveMaximum": false}}}},
			"remote-sync": ActionSpec{
				Description: "Sync a file to a remote host.",
				Params: map[string]interface{}{
					"title":       "remote-sync",
					"description": "Sync a file to a remote host.",
					"type":        "object",
					"properties": map[string]interface{}{
						"file": map[string]interface{}{
							"description": "The file to send out.",
							"type":        "string",
							"format":      "uri"},
						"remote-uri": map[string]interface{}{
							"description": "The host to sync to.",
							"type":        "string",
							"format":      "uri"},
						"util": map[string]interface{}{
							"description": "The util to perform the sync (rsync or scp.)",
							"type":        "string",
							"enum":        []interface{}{"rsync", "scp"}}},
					"required": []interface{}{"file", "remote-uri"}}}}},
	}, {
		description: "A schema with other keys, e.g. \"definitions\"",
		yaml: `
snapshot:
   description: "Take a snapshot of the database."
   params:
      outfile:
         description: "The file to write out to."
         type: "string"
      compression-quality:
         description: "The compression quality."
         type: "integer"
         minimum: 0
         maximum: 9
         exclusiveMaximum: false
   definitions:
      diskdevice: {}
      something-else: {}
`,
		expectedActions: &Actions{map[string]ActionSpec{
			"snapshot": ActionSpec{
				Description: "Take a snapshot of the database.",
				Params: map[string]interface{}{
					"title":       "snapshot",
					"description": "Take a snapshot of the database.",
					"type":        "object",
					"properties": map[string]interface{}{
						"outfile": map[string]interface{}{
							"description": "The file to write out to.",
							"type":        "string",
						},
						"compression-quality": map[string]interface{}{
							"description":      "The compression quality.",
							"type":             "integer",
							"minimum":          0,
							"maximum":          9,
							"exclusiveMaximum": false,
						},
					},
					"definitions": map[string]interface{}{
						"diskdevice":     map[string]interface{}{},
						"something-else": map[string]interface{}{},
					},
				},
			},
		}},
	}, {
		description: "A schema with no \"params\" key, implying no options.",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
`,

		expectedActions: &Actions{map[string]ActionSpec{
			"snapshot": ActionSpec{
				Description: "Take a snapshot of the database.",
				Params: map[string]interface{}{
					"description": "Take a snapshot of the database.",
					"title":       "snapshot",
					"type":        "object",
					"properties":  map[string]interface{}{},
				}}}},
	}, {
		description: "A schema with no values at all, implying no options.",
		yaml: `
snapshot:
`,

		expectedActions: &Actions{map[string]ActionSpec{
			"snapshot": ActionSpec{
				Description: "No description",
				Params: map[string]interface{}{
					"description": "No description",
					"title":       "snapshot",
					"type":        "object",
					"properties":  map[string]interface{}{},
				}}}},
	}}

	// Beginning of testing loop
	for i, test := range goodActionsYamlTests {
		c.Logf("test %d: %s", i, test.description)
		reader := bytes.NewReader([]byte(test.yaml))
		loadedAction, err := ReadActionsYaml(reader)
		c.Assert(err, gc.IsNil)
		c.Check(loadedAction, jc.DeepEquals, test.expectedActions)
	}
}

func (s *ActionsSuite) TestReadBadActionsYaml(c *gc.C) {

	var badActionsYamlTests = []struct {
		description   string
		yaml          string
		expectedError string
	}{{
		description: "Reject JSON-Schema containing references.",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params:
      $schema: "http://json-schema.org/draft-03/schema#"
`,
		expectedError: "schema key \"$schema\" not compatible with this version of juju",
	}, {
		description: "Reject JSON-Schema containing references.",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params:
      outfile: { $ref: "http://json-schema.org/draft-03/schema#" }
`,
		expectedError: "schema key \"$ref\" not compatible with this version of juju",
	}, {
		description: "Malformed YAML: missing key in \"outfile\".",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params:
      outfile:
         The file to write out to.
         type: string
         default: foo.bz2
`,

		expectedError: "YAML error: line 6: mapping values are not allowed in this context",
	}, {
		description: "Malformed JSON-Schema: $schema element misplaced.",
		yaml: `
snapshot:
description: Take a snapshot of the database.
   params:
      outfile:
         $schema: http://json-schema.org/draft-03/schema#
         description: The file to write out to.
         type: string
         default: foo.bz2
`,

		expectedError: "YAML error: line 3: mapping values are not allowed in this context",
	}, {
		description: "Malformed Actions: hyphen at beginning of action name.",
		yaml: `
-snapshot:
   description: Take a snapshot of the database.
`,

		expectedError: "bad action name -snapshot",
	}, {
		description: "Malformed Actions: hyphen after action name.",
		yaml: `
snapshot-:
   description: Take a snapshot of the database.
`,

		expectedError: "bad action name snapshot-",
	}, {
		description: "Malformed Actions: caps in action name.",
		yaml: `
Snapshot:
   description: Take a snapshot of the database.
`,

		expectedError: "bad action name Snapshot",
	}, {
		description: "A non-string description fails to parse",
		yaml: `
snapshot:
   description: ["Take a snapshot of the database."]
`,
		expectedError: "value for schema key \"description\" must be a string",
	}, {
		description: "A non-list \"required\" key",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params:
      outfile:
         description: "The file to write out to."
         type: string
   required: "outfile"
`,
		expectedError: "value for schema key \"required\" must be a YAML list",
	}, {
		description: "A schema with an empty \"params\" key fails to parse",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params:
`,
		expectedError: "params failed to parse as a map",
	}, {
		description: "A schema with a non-map \"params\" value fails to parse",
		yaml: `
snapshot:
   description: Take a snapshot of the database.
   params: ["a", "b"]
`,
		expectedError: "params failed to parse as a map",
	}, {
		description: "\"definitions\" goes against JSON-Schema definition",
		yaml: `
snapshot:
   description: "Take a snapshot of the database."
   params:
      outfile:
         description: "The file to write out to."
         type: "string"
   definitions:
      diskdevice: ["a"]
      something-else: {"a": "b"}
`,
		expectedError: "invalid params schema for action schema snapshot: definitions must be of type array of schemas",
	}, {
		description: "excess keys not in the JSON-Schema spec will be rejected",
		yaml: `
snapshot:
   description: "Take a snapshot of the database."
   params:
      outfile:
         description: "The file to write out to."
         type: "string"
      compression-quality:
         description: "The compression quality."
         type: "integer"
         minimum: 0
         maximum: 9
         exclusiveMaximum: false
   definitions:
      diskdevice: {}
      something-else: {}
   other-key: ["some", "values"],
`,
		expectedError: "YAML error: line 16: did not find expected key",
	}}

	for i, test := range badActionsYamlTests {
		c.Logf("test %d: %s", i, test.description)
		reader := bytes.NewReader([]byte(test.yaml))
		_, err := ReadActionsYaml(reader)
		c.Assert(err, gc.NotNil)
		c.Check(err.Error(), gc.Equals, test.expectedError)
	}
}

func (s *ActionsSuite) TestRecurseMapOnKeys(c *gc.C) {
	tests := []struct {
		should     string
		givenKeys  []string
		givenMap   map[string]interface{}
		expected   interface{}
		shouldFail bool
	}{{
		should:    "fail if the specified key was not in the map",
		givenKeys: []string{"key", "key2"},
		givenMap: map[string]interface{}{
			"key": map[string]interface{}{
				"key": "value",
			},
		},
		shouldFail: true,
	}, {
		should:    "fail if a key was not a string",
		givenKeys: []string{"key", "key2"},
		givenMap: map[string]interface{}{
			"key": map[interface{}]interface{}{
				5: "value",
			},
		},
		shouldFail: true,
	}, {
		should:    "fail if we have more keys but not a recursable val",
		givenKeys: []string{"key", "key2"},
		givenMap: map[string]interface{}{
			"key": []string{"a", "b", "c"},
		},
		shouldFail: true,
	}, {
		should:    "retrieve a good value",
		givenKeys: []string{"key", "key2"},
		givenMap: map[string]interface{}{
			"key": map[string]interface{}{
				"key2": "value",
			},
		},
		expected: "value",
	}, {
		should:    "retrieve a map",
		givenKeys: []string{"key"},
		givenMap: map[string]interface{}{
			"key": map[string]interface{}{
				"key": "value",
			},
		},
		expected: map[string]interface{}{
			"key": "value",
		},
	}, {
		should:    "retrieve a slice",
		givenKeys: []string{"key"},
		givenMap: map[string]interface{}{
			"key": []string{"a", "b", "c"},
		},
		expected: []string{"a", "b", "c"},
	}}

	for i, t := range tests {
		c.Logf("test %d: should %s\n  map: %#v\n  keys: %#v", i, t.should, t.givenMap, t.givenKeys)
		obtained, failed := recurseMapOnKeys(t.givenKeys, t.givenMap)
		c.Assert(!failed, gc.Equals, t.shouldFail)
		if !t.shouldFail {
			c.Check(obtained, jc.DeepEquals, t.expected)
		}
	}
}
