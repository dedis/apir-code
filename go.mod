module github.com/si-co/vpir-code

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dimakogan/dpf-go v0.0.0-20210127221207-b1d9b62bab9b
	github.com/golang/protobuf v1.4.3
	github.com/lukechampine/fastxor v0.0.0-20200124170337-07dbf569dfe7
	github.com/nikirill/go-crypto v0.0.0-20210204153324-694bf46cc691
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210301091718-77cc2087c03b
	golang.org/x/tools v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20210226172003-ab064af71705 // indirect
	google.golang.org/grpc v1.36.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.1 // indirect
	google.golang.org/protobuf v1.25.0
)

//replace github.com/nikirill/go-crypto => ../../go/src/github.com/nikirill/go-crypto
