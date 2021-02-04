module github.com/si-co/vpir-code

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/consensys/goff v0.3.9 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dimakogan/dpf-go v0.0.0-20210127221207-b1d9b62bab9b
	github.com/golang/protobuf v1.4.3
	github.com/lukechampine/fastxor v0.0.0-20200124170337-07dbf569dfe7
	github.com/nikirill/go-crypto v0.0.0-20210204115809-99e94b373076
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c
	golang.org/x/tools v0.0.0-20210105164027-a548c3f4af2d // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20210202153253-cf70463f6119 // indirect
	google.golang.org/grpc v1.35.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.1 // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

//replace github.com/nikirill/go-crypto => ../../go/src/github.com/nikirill/go-crypto
