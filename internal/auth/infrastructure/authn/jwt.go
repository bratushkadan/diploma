package authn

import "github.com/bratushkadan/floral/pkg/auth"

func loadPem(privateKeyPath string, publicKeyPath string) error {
	return nil
}

func foo() {
	_, err := auth.NewJwtProviderBuilder().
		WithPrivateKey(nil).
		WithPublicKey(nil).
		Build()
	if err != nil {
		return
	}
}
