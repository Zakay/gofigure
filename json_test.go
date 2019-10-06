package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

type MarshalJSONTestCase struct {
	data     string
	expected string
}

func TestJsonMarshalCases(t *testing.T) {
	parser := BuildParser()

	testCases := []MarshalJSONTestCase{
		MarshalJSONTestCase{
			data: `
			[test.%{dev,prod}]
				key:"value"`,
			expected: `
			{
				"test" : {
					 "dev" : {
							"key" : "value"
					 },
					 "prod" : {
							"key" : "value"
					 }
				}
		  }`,
		},

		MarshalJSONTestCase{
			data: `
			[test] 
				key:"value" 
			[@.key] 
				key:"value"`,
			expected: `
			{
				"test" : {
					 "key" : {
							"key" : "value"
					 }
				}
		  }`,
		},

		MarshalJSONTestCase{
			data: `
			[%{root1,root2}] 
				key:"value1" 
			[@.%{dev,prod,qa}] 
				key:"value2"`,
			expected: `
			{
				"root1": {
					"key": "value1",
					"dev": {
						"key": "value2"
					},
					"prod": {
						"key": "value2"
					},
					"qa": {
						"key": "value2"
					}
				},
				"root2": {
					"key": "value1",
					"dev": {
						"key":"value2"
					},
					"prod": {
						"key": "value2"
					},
					"qa": {
						"key": "value2"
					}
				}
			}`,
		},

		MarshalJSONTestCase{
			data:     `[root] key:"value" [@.%{dev,prod}.test.%{sub1,sub2}] key:"value"`,
			expected: `{"root":{"key":"value","dev":{"test":{"sub1":{"key":"value"},"sub2":{"key":"value"}}},"prod":{"test":{"sub1":{"key":"value"},"sub2":{"key":"value"}}}}}`,
		},

		/* 	== config.special.fig ==
		# Wut does this mean??
		# special_case_root_key: This key will not be in the same level as dev and production keys

		# Empty value
		empty_value

		# Espace any special characters
		quote_value: \"
		array_value: \[
		*/
		// Maybe do something to ensure determinism here
		MarshalJSONTestCase{
			data:     `Rick_Astley:"Never" \":"gonna" \[:"give" %include "files/config.special.fig"`,
			expected: `{"Rick_Astley":"Never","\\\"":"gonna","\\[":"give","quote_value":"gonna","array_value":"give","empty_value":null}`,
		},

		MarshalJSONTestCase{
			data:     `[root, root2 root3] key:"value"`,
			expected: `{"root":{"key":"value"},"root2":{"key":"value"},"root3":{"key":"value"}}`,
		},
	}

	for _, testCase := range testCases {
		config := &FigureConfig{}

		parser.ParseString(testCase.data, config)

		mappedConfig := config.Transform()

		marshaled, _ := json.Marshal(mappedConfig)

		actual := map[string]interface{}{}
		expected := map[string]interface{}{}

		err := json.Unmarshal(marshaled, &actual)
		if err != nil {
			t.Errorf(err.Error())
		}

		err = json.Unmarshal([]byte(testCase.expected), &expected)
		if err != nil {
			t.Errorf(err.Error())
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\nGot: %s\nExpected: %s\nFrom:%s", string(marshaled), testCase.expected, testCase.data)
		}
	}
}
