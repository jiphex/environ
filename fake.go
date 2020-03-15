package environ

const (
	UnsetEnvPlaceholder = "UNSET!"
)

// fakeEnv includes a function which satisfies lookupableEnvironmentFunc and is
// used in tests within this package to stand in for os.LookupEnv
type fakeEnv struct {
	vals map[string]string
}

func (f fakeEnv) lookupEnv(k string) (string, bool) {
	if f.vals[k] == UnsetEnvPlaceholder {
		return "", false
	}
	return f.vals[k], true
}

// FakeLookupEnv is an implementation of the LookupEnvironmentFunc function
// based around a map[string]string.
//
// Because it's not possible to tell the difference between an unset key and a
// zero-length value in a Go map, there exists a special marker value (stored)
// in UnsetEnvPlaceholder which marks the value as being unset in the
// environment.
func FakeLookupEnv(input map[string]string) LookupEnvironmentFunc {
	f := fakeEnv{vals: input}
	return f.lookupEnv
}

// FakeEmptyEnvironment is an implementation of LookupEnvironmentFunc which
// always pretends that the requested value was not set in the environment.
func FakeEmptyEnvironment() LookupEnvironmentFunc {
	return func(string) (string,bool) {
		return "",false
	}
}