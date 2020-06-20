package utils

import (
	"fmt"
	"testing"
	"github.com/Olling/Enrolld/utils/objects"
)

func TestStructToJson (t *testing.T) {
	var tests = []struct {
		input interface{}
		expected string
	}{
		{
			input: "test",
			expected: "\"test\"",
		},
		{
			input: 1234,
			expected: "1234",
		},
		{
			input: struct {Key string; Value string}{Key: "key", Value: "value"},
			expected: "{\n\t\"Key\": \"key\",\n\t\"Value\": \"value\"\n}",
		},
		{
			input: struct {Boolean bool; Integer int}{true, 12},
			expected: "{\n\t\"Boolean\": true,\n\t\"Integer\": 12\n}",
		},
		{
			input: struct {Boolean bool; Integer int}{false, 13},
			expected: "{\n\t\"Boolean\": false,\n\t\"Integer\": 13\n}",
		},
	}

	for _, test := range tests {
		json, err := StructToJson(test.input)
		if err != nil || json != test.expected {
			t.Error(fmt.Sprintf("Test Failed: Input: '%v', Expected: '%s', JSON: '%s', Error: '%v'", test.input, test.expected, json, err))
		}
	}
}

//func TestStructFromJson(t *testing.T) {
//		t.Error("Not made yet")
//}

func TestStringExistsInArray(t *testing.T) {
	var tests = []struct {
		input []string
		key string
		result bool
	}{
		{
			input: []string{"test1", "test2", "test3"},
			key: "test2",
			result: true,
		},
		{
			input: []string{"test1", "test2", "test3"},
			key: "test1",
			result: true,
		},
		{
			input: []string{"test1", "a1b2c3d4e5f6g7", "test3"},
			key: "test1",
			result: true,
		},
		{
			input: []string{"test1", "test2", "test3", "test4", "test5", "test6", "test7", "test8"},
			key: "test7",
			result: true,
		},
		{
			input: []string{"test1", "test 2", "test3", "test 4", "test5", "test  6", "test7", "test8"},
			key: "test",
			result: false,
		},
	}

	for _, test := range tests {
		if StringExistsInArray(test.input, test.key) != test.result  {
			t.Error(fmt.Sprintf("Test Failed: Input: '%v', Key: '%v'", test.input, test.key))
		}
	}
}

//func KeyValueExistsInMap(chart map[string]string, requiredKey string, requiredValue string) bool {
func TestKeyValueExistsInMap(t *testing.T) {
	var tests = []struct {
		input map[string]string
		key string
		value string
		result bool
	}{
		{
			input: map[string]string{"test 1": "1", "test 2": "2","test 3": "3","test5": "5","test 5": "5","test 6": "6"},
			key: "test 2",
			value: "2",
			result: true,
		},
		{
			input: map[string]string{"test 1": "1", "test 2": "2","test 3": "3","test5": "5","test 5": "5","test 6": "6"},
			key: "test 2",
			value: "3",
			result: false,
		},
		{
			input: map[string]string{"test 1": "1", "test 2": "2","test 3": "3","test5": "5","test 5": "5","test 6": "6"},
			key: "test 34",
			value: "2",
			result: false,
		},
	}

	for _, test := range tests {
		if KeyValueExistsInMap(test.input, test.key, test.value) != test.result  {
			t.Error(fmt.Sprintf("Test Failed: Input: '%v', Key: '%v', Value: '%v', Expected Result: '%v'", test.input, test.key, test.value, test.result))
		}
	}
}

func TestValidInput (t *testing.T) {
	var tests = []struct {
		input string
		result bool
	}{
		{
			input: "abc",
			result: true,
		},
		{
			input: ".,",
			result: false,
		},
		{
			input: "abc123.sh",
			result: true,
		},
		{
			input: "123 abc",
			result: false,
		},
	}

	for _, test := range tests {
		if ValidInput(test.input) != test.result {
			t.Error(fmt.Sprintf("Test Failed: Input: '%s', Result: '%v'", test.input, test.result))
		}
	}
}

func TestGetInventoryInJson(t *testing.T) {
	var server1 objects.Server
	server1.ServerID = "1"
	server1.IP = "192.168.1.5"
	server1.LastSeen = "2020-02-20 20:00:00.423709525 +0100 CET m=+1935960.050243285"
	server1.NewServer = false
	server1.Groups = []string{"group1", "group2"}
	server1.Properties = map[string]string{"test 1": "1", "test 2": "2","test 3": "3","test 6": "6"}
	var server2 objects.Server
	server2.ServerID = "2"
	server2.IP = "192.168.0.7"
	server2.LastSeen = "2018-02-20 20:00:00.423709525 +0100 CET m=+1935960.050243285"
	server2.NewServer = true
	server2.Groups = []string{"group3", "group4"}
	server2.Properties = map[string]string{"something 1": "1", "test 2": "2","test 3": "3","SomethingElse": "5","test 5": "5","test 6": "6"}
	var server3 objects.Server
	server3.ServerID = "3"
	server3.IP = "192.168.10.5"
	server3.LastSeen = "2010-04-20 20:00:00.423709525 +0100 CET m=+1935960.050243285"
	server3.NewServer = false
	server3.Groups = []string{"group5", "group4"}
	server3.Properties =  map[string]string{"something 1": "1"}
	var server4 objects.Server
	server4.ServerID = "4"
	server4.IP = "192.168.10.100"
	server4.LastSeen = "2020-04-20 20:00:00.423709525 +0100 CET m=+1935960.050243285"
	server4.NewServer = true
	server4.Groups = []string{"group5", "group1"}
	server4.Properties =  map[string]string{"What does it cost": "40"}
	var server5 objects.Server
	server5.ServerID = "5"
	server5.IP = "127.0.1.1"
	server5.LastSeen = "2020-06-20 20:00:00.423709525 +0100 CET m=+1935960.050243285"
	server5.NewServer = false
	server5.Groups = []string{"group1"}
	server5.Properties =  map[string]string{"Meaning of life": "42"}

	var tests = []struct {
		input []objects.Server
		result bool
	}{
		{
			input: []objects.Server{server1,server2},
		},
		{
			input: []objects.Server{server5,server2},
		},
		{
			input: []objects.Server{server4,server3,server5},
		},
		{
			input: []objects.Server{server1,server4},
		},
		{
			input: []objects.Server{server4,server3,server2},
		},
		{
			input: []objects.Server{server2,server1},
		},
	}

	for _, test := range tests {
		_, err := GetInventoryInJSON(test.input)

		if err != nil {
			t.Error(fmt.Sprintf("Test Failed: Input: '%v', Error: '%v'", test.input, err))
		}
	}
}
//func VerifyFQDN(serverid string, requestIP string) (string, error) {
