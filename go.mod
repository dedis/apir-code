module github.com/si-co/vpir-code

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cbergoon/merkletree v0.2.0 // indirect
	github.com/consensys/goff v0.3.9 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dimakogan/dpf-go v0.0.0-20210127221207-b1d9b62bab9b
	github.com/golang/protobuf v1.4.3
	github.com/lukechampine/fastxor v0.0.0-20200124170337-07dbf569dfe7
	github.com/nikirill/go-crypto v0.0.0-20210204153324-694bf46cc691
	github.com/stretchr/testify v1.7.0
	github.com/wealdtech/go-merkletree v1.0.0 // indirect
	gitlab.com/NebulousLabs/merkletree v0.0.0-20200118113624-07fbf710afc4 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c
	golang.org/x/tools v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20210203152818-3206188e46ba // indirect
	google.golang.org/grpc v1.35.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.1 // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

//replace github.com/nikirill/go-crypto => ../../go/src/github.com/nikirill/go-crypto
