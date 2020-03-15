package environ

import (
	"reflect"
	"testing"
)

func TestBasicDeepEqual(t *testing.T) {
	a := TestEnv{
		Ghi: false,
	}
	b := TestEnv{
		Ghi: false,
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("structs weren't equal")
	}
}

type TestEnv struct {
	Abc string   `environ:"ABC"`
	Def string   `environ:"DEF,redact"`
	Ghi bool     `environ:"GHI"`
	Jkl bool     `environ:"JKL,anyvaltrue"`
	Mno int      `environ:"MNO"`
	Pqr []string `environ:"PQR"`
}

func TestUnmarshalEnvironment(t *testing.T) {
	type args struct {
		lookupenv LookupEnvironmentFunc
	}
	tests := []struct {
		name    string
		args    args
		want    TestEnv
		wantErr bool
	}{
		{
			name: "does basic stuff",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"JKL": "test",
					"ABC": "abcdef",
					"MNO": "1234",
				}),
			},
			want: TestEnv{
				Abc: "abcdef",
				Jkl: true,
				Mno: 1234,
			},
		},
		{
			name: "has negative project ID",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"JKL": "test",
					"ABC": "abcdef",
					"MNO": "-1234",
				}),
			},
			want: TestEnv{
				Abc: "abcdef",
				Jkl: true,
				Mno: -1234,
			},
		},
		{
			name: "sets non-anyvaltrue bool",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"GHI": "true",
					"JKL": UnsetEnvPlaceholder,
				}),
			},
			want: TestEnv{
				Ghi: true,
			},
		},
		{
			name: "sets non-anyvaltrue bool with empty value",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"GHI": "",
					"JKL": UnsetEnvPlaceholder,
				}),
			},
			want: TestEnv{
				Ghi: false,
			},
		},
		{
			name: "has bad bool value",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"GHI": "xtrue",
					"JKL": UnsetEnvPlaceholder,
				}),
			},
			want: TestEnv{
				Ghi: false,
			},
			wantErr: true,
		},
		{
			name: "jkl not set",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"JKL": UnsetEnvPlaceholder,
				}),
			},
			want: TestEnv{
				Jkl: false,
			},
		},
		{
			name: "empty",
			args: args{
				lookupenv: FakeEmptyEnvironment(),
			},
			want: TestEnv{},
		},
		{
			name: "sets tags to single",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"PQR": "testtag",
					"JKL": UnsetEnvPlaceholder,
				}),
			},
			want: TestEnv{
				Pqr: []string{"testtag"},
			},
		},
		{
			name: "sets multiple tags",
			args: args{
				lookupenv: FakeLookupEnv(map[string]string{
					"PQR": "testtag,othertag,thirdtag",
					"JKL": UnsetEnvPlaceholder,
				}),
			},
			want: TestEnv{
				Pqr: []string{"testtag", "othertag", "thirdtag"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TestEnv{}
			err := UnmarshalEnvironment(tt.args.lookupenv, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalEnvironment() = \n%+v, want \n%+v", got, tt.want)
			}
		})
	}
}
