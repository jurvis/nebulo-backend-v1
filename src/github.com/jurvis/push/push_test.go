package push

import (
	"testing"
)

func TestPushAPNS(t *testing.T) {
	type testset struct {
		alert string
		uuid  string
	}

	var tests = []testset{
		{"The Air is Bad!", "a9777ac5c94b7a69703a67c3c6a93bd541660229e27c81de0c76f57bbfabbf58"},
	}

	for _, pair := range tests {
		v := PushAPNS(pair.alert, pair.uuid)
		if v != nil {
			t.Error(
				"Pushed", pair.alert,
				"to", pair.uuid,
				"push failed.",
			)
		}
	}
}
