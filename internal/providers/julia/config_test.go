package julia

import (
	"reflect"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	tests := map[string]struct {
		env  map[string]string
		want Config
	}{
		"all unset uses defaults": {
			env:  map[string]string{},
			want: Config{Model: DefaultModel, Timeout: DefaultTimeout},
		},
		"model set": {
			env:  map[string]string{"JULIA_MODEL": "julia-risk"},
			want: Config{Model: "julia-risk", Timeout: DefaultTimeout},
		},
		"inference command split": {
			env: map[string]string{"JULIA_INFERENCE_CMD": "python infer.py --model m.onnx"},
			want: Config{
				Model:        DefaultModel,
				InferenceCmd: []string{"python", "infer.py", "--model", "m.onnx"},
				Timeout:      DefaultTimeout,
			},
		},
		"blank inference command unavailable": {
			env:  map[string]string{"JULIA_INFERENCE_CMD": "   "},
			want: Config{Model: DefaultModel, Timeout: DefaultTimeout},
		},
		"model path set": {
			env:  map[string]string{"JULIA_MODEL_PATH": "/models/julia.onnx"},
			want: Config{Model: DefaultModel, ModelPath: "/models/julia.onnx", Timeout: DefaultTimeout},
		},
		"timeout set": {
			env:  map[string]string{"JULIA_TIMEOUT": "5s"},
			want: Config{Model: DefaultModel, Timeout: 5 * time.Second},
		},
		"invalid timeout falls back": {
			env:  map[string]string{"JULIA_TIMEOUT": "not-a-duration"},
			want: Config{Model: DefaultModel, Timeout: DefaultTimeout},
		},
		"non-positive timeout falls back": {
			env:  map[string]string{"JULIA_TIMEOUT": "-5s"},
			want: Config{Model: DefaultModel, Timeout: DefaultTimeout},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}
			got := ConfigFromEnv()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigFromEnv() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigWithDefaults(t *testing.T) {
	tests := map[string]struct {
		in   Config
		want Config
	}{
		"zero value": {
			in:   Config{},
			want: Config{Model: DefaultModel, Timeout: DefaultTimeout},
		},
		"preserves set values": {
			in:   Config{Model: "custom", ModelPath: "/m.onnx", InferenceCmd: []string{"infer"}, Timeout: 3 * time.Second},
			want: Config{Model: "custom", ModelPath: "/m.onnx", InferenceCmd: []string{"infer"}, Timeout: 3 * time.Second},
		},
		"blank model and non-positive timeout": {
			in:   Config{Model: "  ", Timeout: -1},
			want: Config{Model: DefaultModel, Timeout: DefaultTimeout},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.in.WithDefaults()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDefaults() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigWithDefaultsDoesNotMutateReceiver(t *testing.T) {
	orig := Config{}
	_ = orig.WithDefaults()
	if orig.Model != "" || orig.Timeout != 0 {
		t.Errorf("WithDefaults mutated receiver: %+v", orig)
	}
}
