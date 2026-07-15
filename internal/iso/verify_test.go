package iso

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeKeyID(t *testing.T) {
	require.Equal(t, "D94AA3F0EFE21092", normalizeKeyID("0xD94AA3F0EFE21092"))
	require.Equal(t, "843938DF228D22F7B3742BC0D94AA3F0EFE21092", normalizeKeyID("843938DF228D22F7B3742BC0D94AA3F0EFE21092"))
}

func TestInterpretGPGResult(t *testing.T) {
	require.NoError(t, interpretGPGResult("[GNUPG:] GOODSIG abc", "gpg: Good signature", nil))
	require.NoError(t, interpretGPGResult("", "gpg: Poprawny podpis od \"Ubuntu\"", nil))
	require.Error(t, interpretGPGResult("[GNUPG:] BADSIG abc", "", nil))
	require.Error(t, interpretGPGResult("", "gpg: BAD signature", nil))
	err := interpretGPGResult("", "", errors.New("exit status 2"))
	require.Error(t, err)
}

func TestLooksLikePGPKey(t *testing.T) {
	require.True(t, looksLikePGPKey([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n")))
	require.False(t, looksLikePGPKey([]byte("<html>not a key</html>")))
}
