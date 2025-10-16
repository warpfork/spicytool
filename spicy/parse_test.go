package spicy

import (
	"encoding/base64"
	"testing"

	"github.com/go-quicktest/qt"
	"golang.org/x/mod/sumdb/note"
)

const (
	dummyPublicKey = "example.com/log/testdata+33d7b496+AeHTu4Q3hEIMHNqc6fASMsq3rKNx280NI+oO5xCFkkSx"
)

const spicy1 = `c2sp.org/spicy-signature@v1
index 100252
+4IF875THRSj6IOZo2SRpgL6buVI/cawB2iLGFLgWDQ=
nBeqO1BayYMNXX05jMGGtCfprprlSVASVTUUwA0OC9E=
zpUmeEe94VcV/jft8wYlUcDx9yr2kff2jq9QW295vhw=
DTWvScrZyztZdQo5YhJ2Sioo6xDCRDgnoxJJJhN5vl0=
550NhxNMkqzfWFptIePTfVic0rHovJMectLfzaLgh5U=
aoGmNLa5zDVavVfeWhrAxU3duQ0m2TB489dVEGmquO8=
Sbbxxj/8zvCs1q23zn+3wxE9OfJyqVtajvLlvg4FOaI=
WuiUqI+eMKiZ4OUa9CORDO64u91gs7kbTzQYw/e/Khg=
nxkvRAG/1vChJgwuphcHSaCBWcBXebV52Nd3sGxP2kI=
8xVjSf6ZkYSabrgN8m3ePiWgQSpafmz8FXTCfv7guck=
uIkQsFgLVnAeDt5KgZP6l3eJ6YzFz2p92KwlA50Ea5Q=
8SliMOQ4QRiNiRzAt4iEDvrOC0ETwRssbnOdutGqPhA=

example.com/log/testdata
100270
GRV1O0dCqmjjPvYi7CtOT79zb/fPwLNsMFWsynwU8Ac=

— example.com/log/testdata M9e0lkcvy5bAe7rrC8tCyAP+my+CLzswA91zOpBNWVU7qCojkUOHnfdQzDImoYeYH8G8PJSwOqsx0RkSa261OmKVwwM=
`

const spicy2 = spicy1 + `
contexthint
hello there
`

func TestParseSpicySig(t *testing.T) {
	verifier, err := note.NewVerifier(dummyPublicKey)
	if err != nil {
		panic(err)
	}

	t.Run("spicy1", func(t *testing.T) {
		s, err := ParseSpicySig([]byte(spicy1), note.VerifierList(verifier))
		qt.Assert(t, qt.IsNil(err))
		qt.Check(t, qt.Equals(s.entryIndex, 100252))
		qt.Assert(t, qt.HasLen(s.mip, 12))
		qt.Check(t, qt.Equals(base64.StdEncoding.EncodeToString([]byte(s.mip[0][:])), "+4IF875THRSj6IOZo2SRpgL6buVI/cawB2iLGFLgWDQ="))
		qt.Check(t, qt.Equals(base64.StdEncoding.EncodeToString([]byte(s.mip[11][:])), "8SliMOQ4QRiNiRzAt4iEDvrOC0ETwRssbnOdutGqPhA="))
	})

	t.Run("spicy2", func(t *testing.T) {
		s, err := ParseSpicySig([]byte(spicy2), note.VerifierList(verifier))
		qt.Assert(t, qt.IsNil(err))
		qt.Check(t, qt.Equals(s.entryIndex, 100252))
		qt.Assert(t, qt.HasLen(s.mip, 12))
		qt.Check(t, qt.Equals(base64.StdEncoding.EncodeToString([]byte(s.mip[0][:])), "+4IF875THRSj6IOZo2SRpgL6buVI/cawB2iLGFLgWDQ="))
		qt.Check(t, qt.Equals(base64.StdEncoding.EncodeToString([]byte(s.mip[11][:])), "8SliMOQ4QRiNiRzAt4iEDvrOC0ETwRssbnOdutGqPhA="))
		qt.Check(t, qt.Equals(s.contextHint, "hello there\n"))
	})
}
