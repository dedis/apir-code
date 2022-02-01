package utils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"

	"google.golang.org/grpc/credentials"
)

// WARNING: DO NOT USE THESE KEYS IN PRODUCTION!

var ServerPublicKeys = [...]string{
	`-----BEGIN CERTIFICATE-----
MIIDXzCCAcegAwIBAgIRALIZdJxy2Tli+KZ5QpBSqZMwDQYJKoZIhvcNAQELBQAw
ZTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMR0wGwYDVQQLDBRzbXNj
b2xvbUBkZWRpczIwMDIwNDEkMCIGA1UEAwwbbWtjZXJ0IHNtc2NvbG9tQGRlZGlz
MjAwMjA0MB4XDTIyMDIwMTE1NTkwMFoXDTI0MDUwMTE0NTkwMFowRDEnMCUGA1UE
ChMebWtjZXJ0IGRldmVsb3BtZW50IGNlcnRpZmljYXRlMRkwFwYDVQQLDBBzbXNj
b2xvbUBkZWRpc3BjMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEPkBREtus5jB3
Qhm9Qip9DELcbNAJnPq0Xd1wkzgWJ3inEwGINUgXGhLcAE6TNJujopA6PQyhhKT3
o4vcl2oqtKN2MHQwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMB
MB8GA1UdIwQYMBaAFFHjStsu6ll7TUl9SDsBoL3CyM9MMCwGA1UdEQQlMCOCCWxv
Y2FsaG9zdIcEfwAAAYcQAAAAAAAAAAAAAAAAAAAAATANBgkqhkiG9w0BAQsFAAOC
AYEAU0sJGm4NJsPUT36bAH8bFANDsrxwGA8NK1vclW/cc82da03THy5yUWmWp03b
8eHtx0HUw6RQXJ5jvFz21GOCXKPiMDNJdpSy1SDxb2LWMiCQSnhvNyIvy/DAqgpd
BsIMyDiwRS2T9SGEzMuD8LxRDR2pJd2mFQ1HU1lZXTyKBB/P5hR9VZEnoCzx7nah
gC4TtAizDN6nPsGrm7eDvonSgXDE30HcYe3zgKD3OHXaocU6Z1qDpkTw4H09hR56
BXyQghHKKCUnz2eAV30JWLotz0PHNp/ZSOqBeqX4cQCkj+i7hotdjVntS5ar5gKW
T+GjB0QpjRfSCNLv9vKvg7EdoniDTC/NUdwj+Zwc+BpgqZdte4z09INO7l3EZNJZ
iJExqP7XM+J2xmBo8RGYFAH25Fgjj4aRRhf4m2AVgk94fFcjx6NaW+Ows6Iph7zH
UCTY2ShIDUBJf5HHByxUqvjZGljUhMmA/e1JI0dZ170S0eXIIbJvXUj6OVJohKtp
HNZ8
-----END CERTIFICATE-----`,

	`-----BEGIN CERTIFICATE-----
MIIDXjCCAcagAwIBAgIQM+PHm/2bQe5/EbQY4CZ0ZDANBgkqhkiG9w0BAQsFADBl
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExHTAbBgNVBAsMFHNtc2Nv
bG9tQGRlZGlzMjAwMjA0MSQwIgYDVQQDDBtta2NlcnQgc21zY29sb21AZGVkaXMy
MDAyMDQwHhcNMjIwMjAxMTU1OTQ5WhcNMjQwNTAxMTQ1OTQ5WjBEMScwJQYDVQQK
Ex5ta2NlcnQgZGV2ZWxvcG1lbnQgY2VydGlmaWNhdGUxGTAXBgNVBAsMEHNtc2Nv
bG9tQGRlZGlzcGMwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASTFTFIkpgzit+7
Q1HuDtewqbr6mPG2nq095d5MwhAqrq7fYsNf7E2d4m5nqWCvEXdT1tQujeaeyUgj
lZXglM17o3YwdDAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
HwYDVR0jBBgwFoAUUeNK2y7qWXtNSX1IOwGgvcLIz0wwLAYDVR0RBCUwI4IJbG9j
YWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAAAAABMA0GCSqGSIb3DQEBCwUAA4IB
gQAznHJOIX1xw58S6aEGCixOjw32eV7E3vAkClGU0gs/j0iXJyvCa0fd+QpFumFp
U42vIh7vaJVzq7nSMlcKZMRrwZ1rzD8do/QXl907LgAt0sqngQDewxqvWp/lIlAO
7u9lwniCZcmUSDWY0sQcUvOolcKNi3wmdm2k0EbV9YPjNw3PWdlb411h+zo0Ssta
5nJekTpwbaXgFAVMwhjwhWPtqL3yCOjfzh0z7G5fHsmFFXfuAFgNwpcC4l9FhujX
UQvJFd7g/cMQrp456kFWfBIrGR463rdUQVldZY73H+fHfzf8c1gVES61Ojt4fifF
abs4PoNxWToCzl05hKSZrql7c0l2z4qvKE8BMqO34NCFWpr2Ete6mcSJixZQ3a8m
fMX8nYLVchcxrpO4+e8MwyXVizcGWEiBXmymrF+r8Oy1JMg6tdqWRw1Op2KWE7qb
wWyThpwb5w45FmiC2DEDMnpOFfb24vQvM4W79WmstC/H+a/Zaxp3n1OLf2odeveF
z9A=
-----END CERTIFICATE-----`,
}

var serverSecretKeys = [...]string{
	`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg+DRE90ZM2baBKphd
V8CivqnqsYKVjTGHqQXKlTPLgp+hRANCAAQ+QFES26zmMHdCGb1CKn0MQtxs0Amc
+rRd3XCTOBYneKcTAYg1SBcaEtwATpM0m6OikDo9DKGEpPeji9yXaiq0
-----END PRIVATE KEY-----`,

	`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgu1rF5aci2wY6ZFPp
nqqMicePgxvBYh8Mo1Xr7+Ax0XyhRANCAASTFTFIkpgzit+7Q1HuDtewqbr6mPG2
nq095d5MwhAqrq7fYsNf7E2d4m5nqWCvEXdT1tQujeaeyUgjlZXglM17
-----END PRIVATE KEY-----`,
}

// ServerCertificates holds the certificates for the servers
var ServerCertificates []tls.Certificate

func init() {
	numServers := 2

	ServerCertificates = make([]tls.Certificate, numServers)

	var err error
	for i := range ServerCertificates {
		ServerCertificates[i], err = tls.X509KeyPair(
			[]byte(ServerPublicKeys[i]),
			[]byte(serverSecretKeys[i]))
		if err != nil {
			log.Fatalf("could not load certficate #%v %v", i, err)
		}
	}
}

func LoadServersCertificates() (credentials.TransportCredentials, error) {
	cp := x509.NewCertPool()
	for _, cert := range ServerPublicKeys {
		if !cp.AppendCertsFromPEM([]byte(cert)) {
			return nil, errors.New("credentials: failed to append certificates")
		}
	}
	creds := credentials.NewClientTLSFromCert(cp, "127.0.0.1")

	return creds, nil
}
