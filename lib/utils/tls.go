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
MIIDXjCCAcagAwIBAgIQDAxaQih0K9XR2P2ff6Y6EzANBgkqhkiG9w0BAQsFADBl
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExHTAbBgNVBAsMFHNtc2Nv
bG9tQGRlZGlzMjAwMjA0MSQwIgYDVQQDDBtta2NlcnQgc21zY29sb21AZGVkaXMy
MDAyMDQwHhcNMjQwNTE0MDc1NzI3WhcNMjYwODE0MDc1NzI3WjBEMScwJQYDVQQK
Ex5ta2NlcnQgZGV2ZWxvcG1lbnQgY2VydGlmaWNhdGUxGTAXBgNVBAsMEHNtc2Nv
bG9tQGRlZGlzcGMwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARMdXsYkd95AZs1
jqoySUZ4/oOR8cF2wo8qLb57yo5K/7GGF5t4XH00M+G8TR+HDReGQe0fQDLVgDnW
asfJOnyyo3YwdDAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
HwYDVR0jBBgwFoAUUeNK2y7qWXtNSX1IOwGgvcLIz0wwLAYDVR0RBCUwI4IJbG9j
YWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAAAAABMA0GCSqGSIb3DQEBCwUAA4IB
gQBxodQx+6uRGvBOH1FrM9gnbzI53SnYnTzQIGHyPC+vd4fPl6Ohs7sR8mpeQOWy
34GfZCgkFzV30hKA1+hIkbTs9WwcStXIJFNoiZ5UMU5MDSAlqtLDzAyj5vZLs1WF
+9Yq2O0ruDvOltCeS76PPI7gZ1hX5dP3PzC4x5h1xOwg0x+CnAt9UBs5ghdTDCk8
pOo1xvDMslZJ9NIcx8/oO8fWg+Nq6gFeoPvssy7MWHNsHApXRju8B2zc8XH1sptz
V0Ayhwyma9e6oJG+yjnjeGyAPChvuUJjb5JfHS8Ku9Lh1wnYAE/HDjilKczXL02Y
UlPyu+VdM7y9+qpaWYoLaCouO2VKGGVoVzKLKZ8dQWVZZYRefkz/85P0eGzosMSe
U25HF+7GP0TZEfhAlgD9VgIe92xOkEj6rDE2+pHIs4gsAnTTnBY5f7/+Ma10AwF/
CDTtWEy00TU8Aw+xVsrQDl8hD2fWDxJUZ29jvkAfHUPdO/dMXrQgSl8xqJIuBkxq
nvs=
	-----END CERTIFICATE-----`,

	`-----BEGIN CERTIFICATE-----
MIIDXjCCAcagAwIBAgIQDR4dg8xww1hyrEHcmiIzbDANBgkqhkiG9w0BAQsFADBl
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExHTAbBgNVBAsMFHNtc2Nv
bG9tQGRlZGlzMjAwMjA0MSQwIgYDVQQDDBtta2NlcnQgc21zY29sb21AZGVkaXMy
MDAyMDQwHhcNMjQwNTE0MDc1NzUwWhcNMjYwODE0MDc1NzUwWjBEMScwJQYDVQQK
Ex5ta2NlcnQgZGV2ZWxvcG1lbnQgY2VydGlmaWNhdGUxGTAXBgNVBAsMEHNtc2Nv
bG9tQGRlZGlzcGMwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAR/tfC//iEevdK5
wwEYsVul0Hhtu0M8Qz3jCEyvBaE2zm8WwtQ2UHo8nhleFCD+60qcDCqKHGf8vCsu
vpKOIwWko3YwdDAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
HwYDVR0jBBgwFoAUUeNK2y7qWXtNSX1IOwGgvcLIz0wwLAYDVR0RBCUwI4IJbG9j
YWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAAAAABMA0GCSqGSIb3DQEBCwUAA4IB
gQBn64W9fdHTrimXpKWXkpuQlxLuZJYG08spE0nQ8qVjVIy2qxhLHN7SD6Yvnh2h
ry+4eBYwjzLQyJBwbAZX0ZR4nSWYJ+NcOe2CtsVMxIQ1oLE7QifuKIiK0xOqQ83s
IGnvyHiAuv4Bn+x9mgnNprRkMrq5VTwVrRAwXE9rwLHAUmocHkP1VK951icDHLtz
Xnn6NhYQ40P91Mr9vNT3m8TqiEpkGpRJlxS6PexvpCZAJs8bO4Zk8Hk16yhjL087
zTrLlvCmZ8RfxER9bOxDNlpkrQctpZ4B0jIHPJ7QiPllERZH1pvFvQng6QAiS0gC
Ujbd3Pr6XsUcdTYVfcyOKvuMgB28xqP2VxBKwIRWSLb3vN7tWKakqR0jhhBt0pCk
+F+2UOj1aD4qL1EyJRlOyEch9JISMFOPFFeYzPJ1qR3wLFcgyPHhdMxcfapwTbLB
4/DKZVuDvla4Ec/5fERaEK3f5fICk9AYkx2drRMs31NW602pKXcF3ucdMZzTM+cY
dSM=
	-----END CERTIFICATE-----`,
}

var serverSecretKeys = [...]string{
	`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgyHKwZ0LY+KDEQyUL
YHk/PGY2QhmgzlwE1G0NoNpkO1WhRANCAARMdXsYkd95AZs1jqoySUZ4/oOR8cF2
wo8qLb57yo5K/7GGF5t4XH00M+G8TR+HDReGQe0fQDLVgDnWasfJOnyy
	-----END PRIVATE KEY-----`,

	`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgFPH5bH9ShjnSJfzf
3PCbuSeCwXRZ4VxLuYVHITKweBihRANCAAR/tfC//iEevdK5wwEYsVul0Hhtu0M8
Qz3jCEyvBaE2zm8WwtQ2UHo8nhleFCD+60qcDCqKHGf8vCsuvpKOIwWk
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
