// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// matchObjectFromAny returns a map[string]interface{} given any of a filepath,
// an inline YAML string or a map[string]interface{}
func matchObjectFromAny(m interface{}) map[string]interface{} {
	switch m.(type) {
	case string:
		var err error
		var b []byte
		v := m.(string)
		if probablyFilePath(v) {
			b, err = ioutil.ReadFile(v)
			if err != nil {
				// NOTE(jaypipes): We already validated that the file exists at
				// parse time. If we get an error here, just panic cuz there's
				// nothing we can really do.
				panic(err)
			}
		} else {
			b = []byte(v)
		}
		var obj map[string]interface{}
		if err = yaml.Unmarshal(b, &obj); err != nil {
			// NOTE(jaypipes): We already validated that the content could be
			// unmarshaled at parse time. If we get an error here, just panic
			// cuz there's nothing we can really do.
			panic(err)
		}
		return obj
	case map[string]interface{}:
		return m.(map[string]interface{})
	}
	return map[string]interface{}{}
}

type delta struct {
	differences []string
}

func (d *delta) Add(diff string) {
	d.differences = append(d.differences, diff)
}

func (d *delta) Empty() bool {
	return len(d.differences) == 0
}

func (d *delta) Differences() []string {
	return d.differences
}

// compareResourceToMatchObject returns a delta object containing and
// differences between the supplied resource and the match object.
func compareResourceToMatchObject(
	res *unstructured.Unstructured,
	match map[string]interface{},
) *delta {
	d := &delta{differences: []string{}}
	collectFieldDifferences("$", match, res.Object, d)
	return d
}

// collectFieldDifferences compares two things and adds any differences between
// them to a supplied set of differences.
func collectFieldDifferences(
	fp string, // the "field path" to the field we are comparing...
	match interface{},
	subject interface{},
	delta *delta,
) {
	if !typesComparable(match, subject) {
		diff := fmt.Sprintf(
			"%s non-comparable types: %T and %T.",
			fp, match, subject,
		)
		delta.Add(diff)
	}
	switch match.(type) {
	case map[string]interface{}:
		matchmap := match.(map[string]interface{})
		subjectmap := subject.(map[string]interface{})
		for matchk, matchv := range matchmap {
			subjectv, ok := subjectmap[matchk]
			newfp := fp + "." + matchk
			if !ok {
				diff := fmt.Sprintf("%s not present in subject", newfp)
				delta.Add(diff)
			}
			collectFieldDifferences(newfp, matchv, subjectv, delta)
		}
		return
	case []interface{}:
		matchlist := match.([]interface{})
		subjectlist := subject.([]interface{})
		if len(matchlist) != len(subjectlist) {
			diff := fmt.Sprintf(
				"%s had different lengths. expected %d but found %d",
				fp, len(matchlist), len(subjectlist),
			)
			delta.Add(diff)
		}
		// Sort order currently matters, unfortunately...
		for x, matchv := range matchlist {
			subjectv := subjectlist[x]
			newfp := fmt.Sprintf("%s[%d]", fp, x)
			collectFieldDifferences(newfp, matchv, subjectv, delta)
		}
		return
	case int, int8, int16, int32, int64:
		switch subject.(type) {
		case int, int8, int16, int32, int64:
			mv := toInt64(match)
			sv := toInt64(subject)
			if mv != sv {
				diff := fmt.Sprintf(
					"%s had different values. expected %v but found %v",
					fp, match, subject,
				)
				delta.Add(diff)
			}
		case uint, uint8, uint16, uint32, uint64:
			mv := toUint64(match)
			sv := toUint64(subject)
			if mv != sv {
				diff := fmt.Sprintf(
					"%s had different values. expected %v but found %v",
					fp, match, subject,
				)
				delta.Add(diff)
			}
		case string:
			mv := toInt64(match)
			ss := subject.(string)
			sv, err := strconv.Atoi(ss)
			if err != nil {
				diff := fmt.Sprintf(
					"%s had different values. expected %v but found %v",
					fp, match, subject,
				)
				delta.Add(diff)
				return
			}
			if mv != int64(sv) {
				diff := fmt.Sprintf(
					"%s had different values. expected %v but found %v",
					fp, match, subject,
				)
				delta.Add(diff)
			}
		}
		return
	case string:
		switch subject.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			mv := match.(string)
			si := subject.(int)
			sv := strconv.Itoa(si)
			if mv != sv {
				diff := fmt.Sprintf(
					"%s had different values. expected %v but found %v",
					fp, match, subject,
				)
				delta.Add(diff)
			}
		case string:
			mv, _ := match.(string)
			sv, _ := subject.(string)
			if mv != sv {
				diff := fmt.Sprintf(
					"%s had different values. expected %v but found %v",
					fp, match, subject,
				)
				delta.Add(diff)
			}
		}
		return
	}
	if !reflect.DeepEqual(match, subject) {
		diff := fmt.Sprintf(
			"%s had different values. expected %v but found %v",
			fp, match, subject,
		)
		delta.Add(diff)
	}
}

// typesComparable returns true if the two supplied things are comparable,
// false otherwise
func typesComparable(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	at := av.Kind()
	bt := bv.Kind()
	switch at {
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		switch bt {
		case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64,
			reflect.String:
			return true
		default:
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		switch bt {
		case reflect.Uint, reflect.Uint8, reflect.Uint32,
			reflect.Uint64, reflect.String:
			return true
		default:
			return false
		}
	case reflect.Complex64, reflect.Complex128:
		switch bt {
		case reflect.Complex64, reflect.Complex128, reflect.String:
			return true
		default:
			return false
		}
	case reflect.String:
		switch bt {
		case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64,
			reflect.Complex64, reflect.Complex128, reflect.String:
			return true
		default:
			return false
		}
	}
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}

// toUint64 takes an interface and returns a uint64
func toUint64(v interface{}) uint64 {
	switch v.(type) {
	case uint64:
		return v.(uint64)
	case uint8:
		return uint64(v.(uint8))
	case uint16:
		return uint64(v.(uint16))
	case uint32:
		return uint64(v.(uint32))
	case uint:
		return uint64(v.(uint))
	}
	return 0
}

// toInt64 takes an interface and returns an int64
func toInt64(v interface{}) int64 {
	switch v.(type) {
	case int64:
		return v.(int64)
	case int8:
		return int64(v.(int8))
	case int16:
		return int64(v.(int16))
	case int32:
		return int64(v.(int32))
	case int:
		return int64(v.(int))
	}
	return 0
}
