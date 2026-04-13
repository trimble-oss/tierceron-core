package xencryptopts

type Option func(*OptionsBuilder)

type OptionsBuilder struct {
	SetEncryptionSecret func(string) error
	MakeNewEncryption   func() (string, string, error)
	Encrypt             func(input string, encryption map[string]any) (string, error)
	Decrypt             func(passStr string, decryption map[string]any) (string, error)
}

var BuildOptions *OptionsBuilder

func NewOptionsBuilder(opts ...Option) {
	BuildOptions = &OptionsBuilder{}
	for _, opt := range opts {
		opt(BuildOptions)
	}
}
