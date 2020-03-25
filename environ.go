// Package environ is a general libray for taking data from the "environment"
// and parsing it into a struct.
//
// This package should be generally usable outside of the AutoKube ecosystem
// for Go programs which need to load a large number of environment variables
// into a struct.
package environ

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	structFieldTagKey         = "environ"
	structRunnerTagsSeparator = ","
)

// LookupEnvironmentFunc is an interface to allow swapping out of os.LookupEnv
type LookupEnvironmentFunc func(string) (string, bool)

// UnmarshalFromOS just reads the real environment variables and sets the state
func UnmarshalFromOS(into interface{}) error {
	return UnmarshalEnvironment(os.LookupEnv, into)
}

// StateVar is a processed struct tag of a parsed environment variable field
type StateVar struct {
	EnvironmentVariable string
	Redact              bool
	AnyValTrue          bool
}

// parseStateVar is an internal function used to split the Struct tag used in
// this package to allow custom options to be set (e.g redact, anyvaltrue)
func parseStateVar(t reflect.StructTag) (sv StateVar) {
	parts := strings.Split(t.Get(structFieldTagKey), ",")
	sv.EnvironmentVariable = parts[0]
	// panic(t.Get(structFieldTagKey)[0])
	for _, v := range parts[1:] {
		switch v {
		case "redact":
			sv.Redact = true
		case "anyvaltrue":
			sv.AnyValTrue = true
		}
	}
	return
}

// UnmarshalEnvironment does the work of converting environment variables into
// an EnvironmentState using the struct tags on EnvironmentState.
//
// To be recognised by this package, struct fields must be tagged with the
// "environ" tag in the standard Go style
// (FieldName type `environ:"VARNAME,option,option"`).
//
// If a struct field is tagged in this fashion, lookupenv will be queried to
// lookup the environment variable set in VARNAME. The type of the struct field
// will then be checked, and the value of the environment variable will be
// parsed into the struct field value.
//
// In a few situations, UnmarshalEnvironment may fail to set a struct field
// if the parsing of the string environment variable into the native type
// fails. In this case the struct field will be left untouched. One error
// may be returned which would be the last error reached when parsing the
// environment.
//
// Two struct tag options are available to modify the parsing behaviour:
//
// Option "redact" means that the value of the variable will be masked
// when using environ.ToString.
//
// Option "anyvaltrue" means that if the variable is set at all, and of type
// boolean, then the struct field will be set to true (usually the value would
// be parsed with strconv.ParseBool).
func UnmarshalEnvironment(lookupenv LookupEnvironmentFunc, into interface{}) error {
	rv := reflect.ValueOf(into)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("Decode of non-pointer %s", reflect.TypeOf(into))
	}
	if rv.IsNil() {
		return fmt.Errorf("Decode of nil %s", reflect.TypeOf(into))
	}
	st := reflect.TypeOf(into).Elem()
	sv := reflect.ValueOf(into).Elem()
	var err error
	for i := 0; i < st.NumField(); i++ {
		svf := st.Field(i)
		stv := parseStateVar(svf.Tag)
		if svf.Type.Kind() == reflect.Ptr {
			// Need to recurse into pointer
			log.Debugf("from %s recursing into field %d: %s", st, i, svf.Type)
			UnmarshalEnvironment(lookupenv, sv.Field(i).Interface())
		}
		if eval, isset := lookupenv(stv.EnvironmentVariable); isset {
			log.Debugf("environment variable is set: %s", stv.EnvironmentVariable)
			t := svf.Type
			switch {
			case t.Kind() == reflect.Bool:
				if stv.AnyValTrue {
					// Force the value to true because the environment variable is set
					sv.Field(i).SetBool(true)
				} else {
					if len(eval) == 0 {
						log.Tracef("value set with zero length, setting to false")
						sv.Field(i).SetBool(false)
					} else {
						var xb bool
						xb, err = strconv.ParseBool(eval)
						sv.Field(i).SetBool(xb)
					}
				}
			case t.Kind() == reflect.String:
				sv.Field(i).SetString(eval)
			case t == reflect.SliceOf(reflect.TypeOf("")):
				if len(eval) != 0 {
					parts := strings.Split(eval, structRunnerTagsSeparator)
					sv.Field(i).Set(reflect.ValueOf(parts))
				}
			case t.Kind() == reflect.Int:
				if len(eval) == 0 {
					sv.Field(i).SetInt(0)
				} else {
					var xi int
					xi, err = strconv.Atoi(eval)
					sv.Field(i).SetInt(int64(xi))
				}
			case t.Kind() == reflect.Struct:
				panic("struct")
			default:
				log.Fatalf("unimplemented type: %s", t)
			}
		} else {
			log.Tracef("environment variable unset: %s", stv.EnvironmentVariable)
		}
	}
	return err
}

// ToString returns a redacted representation of es
func ToString(es interface{}) string {
	st := reflect.TypeOf(es)
	sb := strings.Builder{}
	sb.WriteString("{ ")
	for i := 0; i < st.NumField(); i++ {
		svf := st.Field(i)
		stv := parseStateVar(svf.Tag)
		val := reflect.ValueOf(es).Field(i).Interface()
		if stv.Redact {
			switch svf.Type.Kind() {
			case reflect.String:
				// val = strings.Repeat("*", len(val.(string)))
				if len(val.(string)) > 0 {
					// I considered this being like strings.Repeat("*", len(val)) but we shouldn't expose the length of the password
					val = "********"
				} else {
					val = ""
				}
			}
		}
		sb.WriteString(fmt.Sprintf(`%s:%v `, svf.Name, val))
	}
	sb.WriteRune('}')
	return sb.String()
}
