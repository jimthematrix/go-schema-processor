package merklize

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-merkletree-sql/v2/db/memory"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDocument = `{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://w3id.org/citizenship/v1",
    "https://w3id.org/security/bbs/v1"
  ],
  "id": "https://issuer.oidp.uscis.gov/credentials/83627465",
  "type": ["VerifiableCredential", "PermanentResidentCard"],
  "issuer": "did:example:489398593",
  "identifier": 83627465,
  "name": "Permanent Resident Card",
  "description": "Government of Example Permanent Resident Card.",
  "issuanceDate": "2019-12-03T12:19:52Z",
  "expirationDate": "2029-12-03T12:19:52Z",
  "credentialSubject": [
    {
      "id": "did:example:b34ca6cd37bbf23",
      "type": ["PermanentResident", "Person"],
      "givenName": "JOHN",
      "familyName": "SMITH",
      "gender": "Male",
      "image": "data:image/png;base64,iVBORw0KGgokJggg==",
      "residentSince": "2015-01-01",
      "lprCategory": "C09",
      "lprNumber": "999-999-999",
      "commuterClassification": "C1",
      "birthCountry": "Bahamas",
      "birthDate": "1958-07-17"
    },
    {
      "id": "did:example:b34ca6cd37bbf24",
      "type": ["PermanentResident", "Person"],
      "givenName": "JOHN",
      "familyName": "SMITH",
      "gender": "Male",
      "image": "data:image/png;base64,iVBORw0KGgokJggg==",
      "residentSince": "2015-01-01",
      "lprCategory": "C09",
      "lprNumber": "999-999-999",
      "commuterClassification": "C1",
      "birthCountry": "Bahamas",
      "birthDate": "1958-07-18"
    }
  ]
}`

func getDataset(t testing.TB, document string) *ld.RDFDataset {
	var obj map[string]interface{}
	err := json.Unmarshal([]byte(document), &obj)
	if err != nil {
		panic(err)
	}

	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Algorithm = "URDNA2015"

	out4, err := proc.Normalize(obj, options)
	require.NoError(t, err)

	out5, ok := out4.(*ld.RDFDataset)
	require.True(t, ok, "%[1]T\n%[1]v", out4)

	return out5
}

func mkPath(parts ...interface{}) Path {
	p, err := NewPath(parts...)
	if err != nil {
		panic(err)
	}
	return p
}

//nolint:deadcode,unused // use for generation of wantEntries
func printEntriesRepresentation(entries []RDFEntry) {
	for _, e := range entries {
		var pathParts []string
		for _, p := range e.key.parts {
			switch p2 := p.(type) {
			case string:
				pathParts = append(pathParts, `"`+p2+`"`)
			case int:
				pathParts = append(pathParts, strconv.Itoa(p2))
			default:
				panic(p)
			}
		}

		var value string
		switch v2 := e.value.(type) {
		case string:
			value = `"` + v2 + `"`
		case int64:
			value = `int64(` + strconv.FormatInt(v2, 10) + `)`
		default:
			panic(fmt.Sprintf("%[1]T -- %[1]v", e.value))
		}
		fmt.Println("{")
		fmt.Printf("key: mkPath(%v),\n", strings.Join(pathParts, ","))
		fmt.Printf("value: %v,\n", value)
		fmt.Println("},")
	}
}

func TestEntriesFromRDF_multigraph(t *testing.T) {
	dataset := getDataset(t, multigraphDoc2)

	entries, err := EntriesFromRDF(dataset)
	require.NoError(t, err)
	// To generate wantEntries, uncomment the following line
	// printEntriesRepresentation(entries)

	wantEntries := []RDFEntry{
		{
			key:   mkPath("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
			value: "https://www.w3.org/2018/credentials#VerifiablePresentation",
		},
		{
			key:   mkPath("https://www.w3.org/2018/credentials#holder", 0),
			value: "http://example.com/holder1",
		},
		{
			key:   mkPath("https://www.w3.org/2018/credentials#holder", 1),
			value: "http://example.com/holder2",
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#verifiableCredential",
				0, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
			value: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld#Iden3SparseMerkleTreeProof",
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#verifiableCredential",
				0,
				"https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#issuerData",
				"https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#state",
				"https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#blockTimestamp"),
			value: int64(123),
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#verifiableCredential",
				1, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
			value: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld#KYCAgeCredential",
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#verifiableCredential",
				1,
				"https://github.com/iden3/claim-schema-vocab/blob/main/credentials/kyc.md#birthday"),
			value: int64(19960424),
		},
	}

	require.Equal(t, wantEntries, entries)
}

func TestEntriesFromRDF(t *testing.T) {
	dataset := getDataset(t, testDocument)

	entries, err := EntriesFromRDF(dataset)
	require.NoError(t, err)

	wantEntries := []RDFEntry{
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://schema.org/birthDate"),
			value: time.Date(1958, 7, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://schema.org/familyName"),
			value: "SMITH",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://schema.org/gender"),
			value: "Male",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://schema.org/givenName"),
			value: "JOHN",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://schema.org/image"),
			value: "data:image/png;base64,iVBORw0KGgokJggg==",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://www.w3.org/1999/02/22-rdf-syntax-ns#type", 0),
			value: "http://schema.org/Person",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"http://www.w3.org/1999/02/22-rdf-syntax-ns#type", 1),
			value: "https://w3id.org/citizenship#PermanentResident",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"https://w3id.org/citizenship#birthCountry"),
			value: "Bahamas",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"https://w3id.org/citizenship#commuterClassification"),
			value: "C1",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"https://w3id.org/citizenship#lprCategory"),
			value: "C09",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"https://w3id.org/citizenship#lprNumber"),
			value: "999-999-999",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 0,
				"https://w3id.org/citizenship#residentSince"),
			value: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://schema.org/birthDate"),
			value: time.Date(1958, 7, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://schema.org/familyName"),
			value: "SMITH",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://schema.org/gender"),
			value: "Male",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://schema.org/givenName"),
			value: "JOHN",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://schema.org/image"),
			value: "data:image/png;base64,iVBORw0KGgokJggg==",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://www.w3.org/1999/02/22-rdf-syntax-ns#type", 0),
			value: "http://schema.org/Person",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"http://www.w3.org/1999/02/22-rdf-syntax-ns#type", 1),
			value: "https://w3id.org/citizenship#PermanentResident",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"https://w3id.org/citizenship#birthCountry"),
			value: "Bahamas",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"https://w3id.org/citizenship#commuterClassification"),
			value: "C1",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"https://w3id.org/citizenship#lprCategory"),
			value: "C09",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"https://w3id.org/citizenship#lprNumber"),
			value: "999-999-999",
		},
		{
			key: mkPath(
				"https://www.w3.org/2018/credentials#credentialSubject", 1,
				"https://w3id.org/citizenship#residentSince"),
			value: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			key:   mkPath("http://schema.org/description"),
			value: "Government of Example Permanent Resident Card.",
		},
		{
			key:   mkPath("http://schema.org/identifier"),
			value: int64(83627465),
		},
		{
			key:   mkPath("http://schema.org/name"),
			value: "Permanent Resident Card",
		},
		{
			key:   mkPath("http://www.w3.org/1999/02/22-rdf-syntax-ns#type", 0),
			value: "https://w3id.org/citizenship#PermanentResidentCard",
		},
		{
			key:   mkPath("http://www.w3.org/1999/02/22-rdf-syntax-ns#type", 1),
			value: "https://www.w3.org/2018/credentials#VerifiableCredential",
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#credentialSubject",
				0),
			value: "did:example:b34ca6cd37bbf23",
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#credentialSubject",
				1),
			value: "did:example:b34ca6cd37bbf24",
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#expirationDate"),
			//value: "2029-12-03T12:19:52Z",
			value: time.Date(2029, 12, 3, 12, 19, 52, 0, time.UTC),
		},
		{
			key: mkPath("https://www.w3.org/2018/credentials#issuanceDate"),
			//value: "2019-12-03T12:19:52Z",
			value: time.Date(2019, 12, 3, 12, 19, 52, 0, time.UTC),
		},
		{
			key:   mkPath("https://www.w3.org/2018/credentials#issuer"),
			value: "did:example:489398593",
		},
	}
	require.Equal(t, wantEntries, entries)
}

func TestProof(t *testing.T) {
	dataset := getDataset(t, testDocument)

	entries, err := EntriesFromRDF(dataset)
	require.NoError(t, err)

	ctx := context.Background()

	mt, err := merkletree.NewMerkleTree(ctx, memory.NewMemoryStorage(), 40)
	require.NoError(t, err)

	err = AddEntriesToMerkleTree(ctx, mt, entries)
	require.NoError(t, err)

	// [https://www.w3.org/2018/credentials#credentialSubject 1 http://schema.org/birthDate] => 1958-07-18
	path, err := NewPath(
		"https://www.w3.org/2018/credentials#credentialSubject", 1,
		"http://schema.org/birthDate")
	require.NoError(t, err)

	birthDate := time.Date(1958, 7, 18, 0, 0, 0, 0, time.UTC)
	entry, err := NewRDFEntry(path, birthDate)
	require.NoError(t, err)

	key, val, err := entry.KeyValueMtEntries()
	require.NoError(t, err)

	p, _, err := mt.GenerateProof(ctx, key, nil)
	require.NoError(t, err)

	ok := merkletree.VerifyProof(mt.Root(), p, key, val)
	require.True(t, ok)
}

func TestProofInteger(t *testing.T) {
	dataset := getDataset(t, testDocument)

	entries, err := EntriesFromRDF(dataset)
	require.NoError(t, err)

	ctx := context.Background()

	mt, err := merkletree.NewMerkleTree(ctx, memory.NewMemoryStorage(), 40)
	require.NoError(t, err)

	err = AddEntriesToMerkleTree(ctx, mt, entries)
	require.NoError(t, err)

	path, err := NewPath("http://schema.org/identifier")
	require.NoError(t, err)

	entry, err := NewRDFEntry(path, 83627465)
	require.NoError(t, err)

	key, val, err := entry.KeyValueMtEntries()
	require.NoError(t, err)

	p, _, err := mt.GenerateProof(ctx, key, nil)
	require.NoError(t, err)

	ok := merkletree.VerifyProof(mt.Root(), p, key, val)
	require.True(t, ok)
}

func TestMerklizer_Proof(t *testing.T) {
	ctx := context.Background()
	mz, err := MerklizeJSONLD(ctx, strings.NewReader(testDocument))
	require.NoError(t, err)

	t.Run("test with path as Path", func(t *testing.T) {
		// [https://www.w3.org/2018/credentials#credentialSubject 1 http://schema.org/birthDate] => 1958-07-18
		path, err := NewPath(
			"https://www.w3.org/2018/credentials#credentialSubject", 1,
			"http://schema.org/birthDate")
		require.NoError(t, err)

		p, value, err := mz.Proof(ctx, path)
		require.NoError(t, err)

		pathMtEntry, err := path.MtEntry()
		require.NoError(t, err)

		require.True(t, value.IsTime())
		valueDateType, err := value.AsTime()
		require.NoError(t, err)

		birthDate := time.Date(1958, 7, 18, 0, 0, 0, 0, time.UTC)
		require.True(t, birthDate.Equal(valueDateType))

		valueMtEntry, err := value.MtEntry()
		require.NoError(t, err)

		ok := merkletree.VerifyProof(mz.Root(), p, pathMtEntry, valueMtEntry)
		require.True(t, ok)
	})

	t.Run("test with path as shortcut string", func(t *testing.T) {
		path, err := mz.ResolveDocPath("credentialSubject.1.birthCountry")
		require.NoError(t, err)

		p, value, err := mz.Proof(ctx, path)
		require.NoError(t, err)

		require.True(t, value.IsString())
		valueStr, err := value.AsString()
		require.NoError(t, err)
		require.Equal(t, "Bahamas", valueStr)
		valueMtEntry, err := value.MtEntry()
		require.NoError(t, err)

		pathMtEntry, err := path.MtEntry()
		require.NoError(t, err)

		ok := merkletree.VerifyProof(mz.Root(), p, pathMtEntry, valueMtEntry)
		require.True(t, ok)
	})

	t.Run("test RawValue", func(t *testing.T) {
		// [https://www.w3.org/2018/credentials#credentialSubject 1 http://schema.org/birthDate] => 1958-07-18
		path, err := NewPath(
			"https://www.w3.org/2018/credentials#credentialSubject", 1,
			"http://schema.org/birthDate")
		require.NoError(t, err)

		// Check RawValue with index in path
		rv, err := mz.RawValue(path)
		require.NoError(t, err)
		require.Equal(t, "1958-07-18", rv)

		// Check RawValue as a number in json
		identifierPath, err := NewPath("http://schema.org/identifier")
		require.NoError(t, err)
		rv, err = mz.RawValue(identifierPath)
		require.NoError(t, err)
		require.Equal(t, float64(83627465), rv)

		// Check RawValue with wrong path (expected array, but got object)
		wrongPath, err := NewPath("http://schema.org/identifier", 1)
		require.NoError(t, err)
		_, err = mz.RawValue(wrongPath)
		require.EqualError(t, err,
			"expected array at 'http://schema.org/identifier / [1]'")

		wrongPath, err = NewPath(
			"https://www.w3.org/2018/credentials#credentialSubject", 5,
			"http://schema.org/birthDate")
		require.NoError(t, err)
		_, err = mz.RawValue(wrongPath)
		require.EqualError(t, err,
			"index is out of range at 'https://www.w3.org/2018/credentials#credentialSubject / [5]'")

		wrongPath, err = NewPath("bla-bla", 5)
		require.NoError(t, err)
		_, err = mz.RawValue(wrongPath)
		require.EqualError(t, err, "value not found at 'bla-bla'")

		wrongPath, err = NewPath(
			"https://www.w3.org/2018/credentials#credentialSubject", "bla-bla")
		require.NoError(t, err)
		_, err = mz.RawValue(wrongPath)
		require.EqualError(t, err,
			"expected object at 'https://www.w3.org/2018/credentials#credentialSubject / bla-bla'")
	})

	mzRoot := mz.Root()
	require.Equal(t,
		"d001de1d1b74d3b24b394566511da50df18532264c473845ea51e915a588b02a",
		mzRoot.Hex())
}

//nolint:deadcode,unused // use for debugging
func logDataset(in *ld.RDFDataset) {
	fmt.Printf("Log dataset of %v keys\n", len(in.Graphs))
	for s, gs := range in.Graphs {
		fmt.Printf("Key %v has %v entries\n", s, len(gs))
		for i, g := range gs {
			subject := "nil"
			if g.Subject != nil {
				subject = g.Subject.GetValue()
			}
			predicate := "nil"
			if g.Predicate != nil {
				predicate = g.Predicate.GetValue()
			}
			object := "nil"
			var ol2 string
			ol, olOK := g.Object.(*ld.Literal)
			if olOK {
				ol2 = ol.Datatype
			}

			if g.Object != nil {
				object = g.Object.GetValue()
			}
			graph := "nil"
			if g.Graph != nil {
				graph = g.Graph.GetValue()
			}
			fmt.Printf(`Entry %v:
	Subject [%T]: %v
	Predicate [%T]: %v
	Object [%T]: %v %v
	Graph [%T]: %v
`, i,
				g.Subject, subject,
				g.Predicate, predicate,
				g.Object, object, ol2,
				g.Graph, graph)
		}
	}
}

//nolint:deadcode,unused //reason: used in debugging
func logEntries(es []RDFEntry) {
	for i, e := range es {
		log.Printf("Entry %v: %v => %v", i, fmtPath(e.key), e.value)
	}
}

//nolint:deadcode,unused //reason: used in debugging
func fmtPath(p Path) string {
	var parts []string
	for _, pi := range p.parts {
		switch v := pi.(type) {
		case string:
			parts = append(parts, v)
		case int:
			parts = append(parts, strconv.Itoa(v))
		default:
			panic("not string or int")
		}
	}
	return strings.Join(parts, " :: ")
}

func TestPathFromContext(t *testing.T) {
	// this file downloaded from here: https://www.w3.org/2018/credentials/v1
	ctxBytes, err := os.ReadFile("testdata/credentials_v1.json")
	require.NoError(t, err)

	in := "VerifiableCredential.credentialSchema.JsonSchemaValidator2018"
	result, err := NewPathFromContext(ctxBytes, in)
	require.NoError(t, err)

	want, err := NewPath(
		"https://www.w3.org/2018/credentials#VerifiableCredential",
		"https://www.w3.org/2018/credentials#credentialSchema",
		"https://www.w3.org/2018/credentials#JsonSchemaValidator2018")
	require.NoError(t, err)

	require.Equal(t, want, result)
}

func TestFieldPathFromContext(t *testing.T) {
	ctxBytes, err := os.ReadFile("testdata/credential_2.json-ld")
	require.NoError(t, err)

	typ := "KYCAgeCredential"
	fieldPath := "birthday"
	result, err := NewFieldPathFromContext(ctxBytes, typ, fieldPath)
	require.NoError(t, err)

	want, err := NewPath(
		"https://github.com/iden3/claim-schema-vocab/blob/main/credentials/kyc.md#birthday")
	require.NoError(t, err)

	require.Equal(t, want, result)
}

func TestPathFromDocument(t *testing.T) {
	in := "credentialSubject.1.birthDate"
	result, err := NewPathFromDocument([]byte(testDocument), in)
	require.NoError(t, err)

	want, err := NewPath(
		"https://www.w3.org/2018/credentials#credentialSubject",
		1,
		"http://schema.org/birthDate")
	require.NoError(t, err)

	require.Equal(t, want, result)
}

func TestMkValueInt(t *testing.T) {
	testCases := []struct {
		in   int64
		want string
	}{
		{
			in:   -1,
			want: "21888242871839275222246405745257275088548364400416034343698204186575808495616",
		},
		{
			in:   -2,
			want: "21888242871839275222246405745257275088548364400416034343698204186575808495615",
		},
		{
			in:   math.MinInt64,
			want: "21888242871839275222246405745257275088548364400416034343688980814538953719809",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("#%v", i+1), func(t *testing.T) {
			v, err := mkValueInt(defaultHasher, tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, v.Text(10))
		})
	}

	t.Run("int value", func(t *testing.T) {
		v, err := mkValueInt(defaultHasher, int(math.MinInt64))
		require.NoError(t, err)
		require.Equal(t,
			"21888242871839275222246405745257275088548364400416034343688980814538953719809",
			v.Text(10))
	})
}

func TestValue(t *testing.T) {
	// bool
	v, err := NewValue(defaultHasher, true)
	require.NoError(t, err)
	require.False(t, v.IsString())
	require.True(t, v.IsBool())
	require.False(t, v.IsInt64())
	require.False(t, v.IsTime())
	b, err := v.AsBool()
	require.NoError(t, err)
	require.True(t, b)
	_, err = v.AsString()
	require.ErrorIs(t, err, ErrIncorrectType)

	// string
	s, err := NewValue(defaultHasher, "str")
	require.NoError(t, err)
	require.True(t, s.IsString())
	require.False(t, s.IsBool())
	require.False(t, s.IsInt64())
	require.False(t, s.IsTime())
	s2, err := s.AsString()
	require.NoError(t, err)
	require.Equal(t, "str", s2)
	_, err = s.AsInt64()
	require.ErrorIs(t, err, ErrIncorrectType)

	// string
	i, err := NewValue(defaultHasher, int64(3))
	require.NoError(t, err)
	require.False(t, i.IsString())
	require.False(t, i.IsBool())
	require.True(t, i.IsInt64())
	require.False(t, i.IsTime())
	i2, err := i.AsInt64()
	require.NoError(t, err)
	require.Equal(t, int64(3), i2)
	_, err = i.AsTime()
	require.ErrorIs(t, err, ErrIncorrectType)

	// time.Time
	tm := time.Date(2022, 10, 20, 3, 4, 5, 6, time.UTC)
	tm2, err := NewValue(defaultHasher, tm)
	require.NoError(t, err)
	require.False(t, tm2.IsString())
	require.False(t, tm2.IsBool())
	require.False(t, tm2.IsInt64())
	require.True(t, tm2.IsTime())
	tm3, err := tm2.AsTime()
	require.NoError(t, err)
	require.True(t, tm3.Equal(tm))
	_, err = tm2.AsBool()
	require.ErrorIs(t, err, ErrIncorrectType)
}

// multiple types within another type
var doc1 = `
{
    "@context": [
        "https://www.w3.org/2018/credentials/v1",
        "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld",
        "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"
    ],
    "@type": [
        "VerifiableCredential",
        "KYCAgeCredential"
    ],
    "id": "http://myid.com",
    "expirationDate": "2361-03-21T21:14:48+02:00",
    "credentialSubject": {
        "type": "KYCAgeCredential",
        "id": "did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ",
        "documentType": 1,
        "birthday": 19960424
    },
    "credentialStatus": {
        "type": "SparseMerkleTreeProof",
        "id": "http://localhost:8001/api/v1/identities/1195DjqzhZ9zpHbezahSevDMcxN41vs3Y6gb4noRW/claims/revocation/status/127366661"
    },
    "credentialSchema": {
        "type": "JsonSchemaValidator2018",
        "id": "http://json1.com"
    }
}`

func TestExistenceProof(t *testing.T) {
	ctx := context.Background()
	mz, err := MerklizeJSONLD(ctx, strings.NewReader(doc1))
	require.NoError(t, err)
	path, err := mz.ResolveDocPath("credentialSubject.birthday")
	require.NoError(t, err)

	wantPath, err := NewPath(
		"https://www.w3.org/2018/credentials#credentialSubject",
		"https://github.com/iden3/claim-schema-vocab/blob/main/credentials/kyc.md#birthday")
	require.NoError(t, err)
	require.Equal(t, wantPath, path)

	p, v, err := mz.Proof(ctx, path)
	require.NoError(t, err)

	require.True(t, p.Existence)
	i, err := v.AsInt64()
	require.NoError(t, err)
	require.Equal(t, int64(19960424), i)
}

func findQuadByObject(t testing.TB, ds *ld.RDFDataset, value any) *ld.Quad {
	for _, quads := range ds.Graphs {
		for _, quad := range quads {
			if reflect.DeepEqual(value, quad.Object) {
				return quad
			}
		}
	}

	t.Fatal("quad not found")
	return nil
}

func findQuadByIdx(t testing.TB, ds *ld.RDFDataset, idx datasetIdx) *ld.Quad {
	quads, ok := ds.Graphs[idx.graph]
	if !ok {
		t.Fatal("graph not found")
	}
	if len(quads) < idx.idx+1 {
		t.Fatal("quad not found")
	}
	return quads[idx.idx]
}

const multigraphDoc2 = `{
  "@context":[
    "https://www.w3.org/2018/credentials/v1",
    "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld",
    "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld"
  ],
  "@type":"VerifiablePresentation",
  "holder": ["http://example.com/holder1", "http://example.com/holder2"],
  "verifiableCredential":[
    {
      "@id": "http://example.com/vc1",
      "@type":"KYCAgeCredential",
      "birthday":19960424
    },
    {
      "@id": "http://example.com/vc3",
      "@type": "Iden3SparseMerkleTreeProof",
      "issuerData": {
        "state": {
          "blockTimestamp": 123
        }
      }
    }
  ]
}`

func Test_findParentInsideGraph_And_findGraphParent(t *testing.T) {
	ds := getDataset(t, multigraphDoc2)
	q := findQuadByObject(t, ds, &ld.Literal{
		Value:    "123",
		Datatype: ld.XSDInteger,
		Language: "",
	})
	idx, err := findParentInsideGraph(ds, q)
	require.NoError(t, err)
	q = findQuadByIdx(t, ds, idx)
	assert.Equal(t,
		&ld.IRI{Value: "https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#state"},
		q.Predicate)

	idx, err = findParentInsideGraph(ds, q)
	require.NoError(t, err)
	q = findQuadByIdx(t, ds, idx)
	assert.Equal(t,
		&ld.IRI{Value: "http://example.com/vc3"},
		q.Subject)
	assert.Equal(t,
		&ld.IRI{Value: "https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#issuerData"},
		q.Predicate)

	_, err = findParentInsideGraph(ds, q)
	require.ErrorIs(t, err, errParentNotFound)

	idx, err = findGraphParent(ds, q)
	require.NoError(t, err)
	q = findQuadByIdx(t, ds, idx)
	assert.Equal(t,
		&ld.IRI{Value: "https://www.w3.org/2018/credentials#verifiableCredential"},
		q.Predicate)

	_, err = findParentInsideGraph(ds, q)
	require.ErrorIs(t, err, errParentNotFound)
}

func Test_findParent(t *testing.T) {
	ds := getDataset(t, multigraphDoc2)
	q := findQuadByObject(t, ds, &ld.Literal{
		Value:    "123",
		Datatype: ld.XSDInteger,
		Language: "",
	})
	idx, err := findParent(ds, q)
	require.NoError(t, err)
	q = findQuadByIdx(t, ds, idx)
	assert.Equal(t,
		&ld.IRI{Value: "https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#state"},
		q.Predicate)

	idx, err = findParent(ds, q)
	require.NoError(t, err)
	q = findQuadByIdx(t, ds, idx)
	assert.Equal(t,
		&ld.IRI{Value: "http://example.com/vc3"},
		q.Subject)
	assert.Equal(t,
		&ld.IRI{Value: "https://github.com/iden3/claim-schema-vocab/blob/main/proofs/Iden3SparseMerkleTreeProof-v2.md#issuerData"},
		q.Predicate)

	idx, err = findParent(ds, q)
	require.NoError(t, err)
	q = findQuadByIdx(t, ds, idx)
	assert.Equal(t,
		&ld.IRI{Value: "https://www.w3.org/2018/credentials#verifiableCredential"},
		q.Predicate)

	_, err = findParent(ds, q)
	require.ErrorIs(t, err, errParentNotFound)
}

const multigraphDoc = `{
  "@context":[
    "https://www.w3.org/2018/credentials/v1",
    "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"
  ],
  "@type":"VerifiablePresentation",
  "holder": ["http://example.com/holder1", "http://example.com/holder2"],
  "verifiableCredential": {
    "@id": "http://example.com/vc2",
    "@type":"KYCAgeCredential",
    "birthday":19960425
  }
}`

func TestMerklizer_RawValue(t *testing.T) {
	ctx := context.Background()
	mz, err := MerklizeJSONLD(ctx, strings.NewReader(multigraphDoc))
	require.NoError(t, err)

	path, err := NewPathFromDocument([]byte(multigraphDoc),
		"verifiableCredential.birthday")
	require.NoError(t, err)

	require.NoError(t, err)

	val, err := mz.RawValue(path)
	require.NoError(t, err)
	require.Equal(t, float64(19960425), val)
}

func TestIncorrectDocument_UnsafeMode(t *testing.T) {
	const docUnknownFields = `{
    "id": "http://127.0.0.1/id",
    "expirationDate": "2030-01-01T00:00:00Z"
}`

	ctx := context.Background()

	t.Run("default safe mode", func(t *testing.T) {
		_, err := MerklizeJSONLD(ctx, strings.NewReader(docUnknownFields))
		require.EqualError(t, err,
			"invalid property: Dropping property that did not expand into an absolute IRI or keyword.")

	})

	t.Run("explicitly set safe mode", func(t *testing.T) {
		_, err := MerklizeJSONLD(ctx, strings.NewReader(docUnknownFields),
			WithSafeMode(true))
		require.EqualError(t, err,
			"invalid property: Dropping property that did not expand into an absolute IRI or keyword.")
	})

	t.Run("explicitly set unsafe mode", func(t *testing.T) {
		_, err := MerklizeJSONLD(ctx, strings.NewReader(docUnknownFields),
			WithSafeMode(false))
		require.NoError(t, err)
	})
}
