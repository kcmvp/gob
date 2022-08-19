package builder

// DefaultHookMsg default git commit message format, which contains 3 parts
// 1: first part is a '#' and follow with a number and its length range is [1,7].
// 2: second part is a separator ':'
// 3: third part is the message body, which minimize length is 10 characters
// you can do the necessary changes according your requirements.
const DefaultHookMsg = `#[0-9]{1,7}:\s?(\S+\s?){10,}`

type HookMsgPattern string

type buildOption struct {
	MinCoverage float64
	MaxCoverage float64
	MsgPattern  HookMsgPattern
}

func defaultOption() *buildOption {
	return &buildOption{
		MinCoverage: 0.35,
		MaxCoverage: 0.90,
		MsgPattern:  DefaultHookMsg,
	}
}
