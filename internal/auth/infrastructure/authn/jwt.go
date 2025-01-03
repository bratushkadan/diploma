package authn

import "github.com/bratushkadan/floral/pkg/auth"

func loadPem(privateKeyPath string, publicKeyPath string) error {
	return nil
}

func foo() {
	conf := auth.JwtProviderConf{
		PrivateKey: []byte{},
		PublicKey:  []byte{},
	}
	auth.NewJwtProvider(conf)
}
