package scaffolds

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateCommitMsg(t *testing.T) {
	tests := []struct {
		name  string
		msg   string
		valid bool
	}{
		{
			"valid01",
			"#123: must be fix by v1",
			true,
		},
		{
			"valid02",
			"#123: must be fix by v1",
			true,
		},
		{
			"invalid01",
			"#123 : must be fix by v1",
			false,
		},
		{
			"valid03",
			"#123:must be fix by v1",
			true,
		},
		{
			"invalid02",
			"#: must be fix by v1",
			false,
		},
		{
			"invalid03",
			"#123l: must be fix by v1",
			false,
		},
		{
			"invalid04",
			"#123: must ",
			false,
		},
		{
			"invalid05",
			"#123#234: fix multiple issues ",
			false,
		},
		{
			"valid04",
			"#123 #234: fix multiple issues ",
			true,
		},
		{
			"valid05",
			"#234: fix multiple issues ",
			true,
		},
		{
			"valid06",
			" #234: fix multiple issues ",
			true,
		},
		{
			"valid07",
			" #234 #12345: fix multiple issues ",
			true,
		},
		{
			"valid08",
			" #234   #12345: fix multiple issues ",
			true,
		},
		{
			"invalid06",
			"#123,#234,: fix multiple issues ",
			false,
		},
		{
			"invalid07",
			",#234: fix multiple issues ",
			false,
		},
		{
			"invalid08",
			"#234 : fix multiple issues ",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCommitMsg(test.msg, DefaultHookMsg)
			require.True(t, (err == nil) == test.valid, func() string {
				if test.valid {
					return "should be a valid message"
				} else {
					return "should be a invalid message"
				}
			}())
		})
	}
}
