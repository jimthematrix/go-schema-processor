module github.com/iden3/go-schema-processor

go 1.18

require (
	github.com/iden3/go-iden3-core v1.0.0
	github.com/iden3/go-iden3-crypto v0.0.13
	github.com/iden3/go-merkletree-sql/v2 v2.0.0
	github.com/ipfs/go-ipfs-api v0.3.0
	// We require the `json-gold` bugfix which has not yet been included in the
	// stable version. After the release of version 0.5.1 or later, it will be
	// necessary to update to the stable version.
	// https://github.com/piprate/json-gold/commit/36fcca9d7e487684a764e552e7d837a14546a157
	github.com/piprate/json-gold v0.5.1-0.20230111113000-6ddbe6e6f19f
	github.com/pkg/errors v0.9.1
	github.com/qri-io/jsonschema v0.2.1
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.0.0-20220126234351-aa10faf2a1f8
)

require (
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20170627025303-887ab5e44cc3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dchest/blake512 v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/ipfs/go-cid v0.0.7 // indirect
	github.com/ipfs/go-ipfs-files v0.0.9 // indirect
	github.com/libp2p/go-buffer-pool v0.0.2 // indirect
	github.com/libp2p/go-flow-metrics v0.0.3 // indirect
	github.com/libp2p/go-libp2p-core v0.6.1 // indirect
	github.com/libp2p/go-openssl v0.0.7 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.0.3 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.0 // indirect
	github.com/multiformats/go-multibase v0.0.3 // indirect
	github.com/multiformats/go-multihash v0.0.14 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/qri-io/jsonpointer v0.1.1 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/whyrusleeping/tar-utils v0.0.0-20180509141711-8c6c8ba81d5c // indirect
	go.opencensus.io v0.22.4 // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
