package snippet

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AppStore struct {
	Issuer      string
	KeyId       string
	KeyFileName string
	PackageName string
}

func (a *AppStore) readPrivateKeyFromFile() (*ecdsa.PrivateKey, error) {
	exec, err := os.Executable()
	if err != nil {
		zenlog.Error("appstore failed to os executable err %+v", err)
		return nil, err
	}

	WorkingDir := filepath.Dir(exec)
	keyFile := fmt.Sprintf("%s/secrets/%v", WorkingDir, a.KeyFileName)
	bytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("appstore private key must be a valid .p8 PEM file")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		return pk, nil
	default:
		return nil, errors.New("appstore private key must be of type ecdsa.PrivateKey")
	}
}

// Per doc: https://developer.apple.com/documentation/appstoreserverapi#topics
func (a *AppStore) Do(method string, url string, body io.Reader) (int, []byte, error) {
	privateKey, err := a.readPrivateKeyFromFile()
	if err != nil {
		zenlog.Error("appstore read private key err %+v", err)
		return 0, nil, err
	}

	// TODO: reuse the same signed token for up to 60 minutes
	authToken, err := a.generateToken(privateKey)
	if err != nil {
		zenlog.Error("appstore generate token err %+v", err)
		return 0, nil, err
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest(
		method,
		url,
		body,
	)
	if err != nil {
		zenlog.Error("appstore new http request err %+v", err)
		return 0, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "App Store Client")
	resp, err := client.Do(req)
	if err != nil {
		zenlog.Error("appstore http client do err %+v", err)
		return 0, nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zenlog.Error("appstore read http body err %+v", err)
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, bytes, err
}

// Per doc: https://developer.apple.com/documentation/appstoreserverapi/generating_tokens_for_api_requests
func (a *AppStore) generateToken(privateKey *ecdsa.PrivateKey) (string, error) {
	expirationTimestamp := time.Now().Add(31 * time.Minute)
	id := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss":   a.Issuer,
		"isa":   time.Now().Unix(),
		"exp":   expirationTimestamp.Unix(),
		"aud":   "appstoreconnect-v1",
		"nonce": id.String(),
		"bid":   a.PackageName,
	})

	token.Header["kid"] = a.KeyId
	token.Header["alg"] = "ES256"
	token.Header["typ"] = "JWT"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// https://developer.apple.com/documentation/appstoreserverapi/jwstransaction?changes=latest_major
type JWSTransaction struct {
	BundleID             string `json:"bundleId"`
	InAppOwnershipType   string `json:"inAppOwnershipType"`
	TransactionID        string `json:"transactionId"`
	ProductID            string `json:"productId"`
	PurchaseDate         int64  `json:"purchaseDate"`
	Type                 string `json:"type"`
	OriginalPurchaseDate int64  `json:"originalPurchaseDate"`
}

func (ac *JWSTransaction) Valid() error {

	return nil
}

// Per doc: https://datatracker.ietf.org/doc/html/rfc7515#section-4.1.6
func (a *AppStore) extractPublicKeyFromToken(tokenStr string) (*ecdsa.PublicKey, error) {
	certStr, err := a.extractHeaderByIndex(tokenStr, 0)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certStr)
	if err != nil {
		return nil, err
	}

	switch pk := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		return pk, nil
	default:
		return nil, errors.New("appstore public key must be of type ecdsa.PublicKey")
	}
}

func (a *AppStore) extractHeaderByIndex(tokenStr string, index int) ([]byte, error) {
	if index > 2 {
		return nil, errors.New("invalid index")
	}

	tokenArr := strings.Split(tokenStr, ".")
	headerByte, err := base64.RawStdEncoding.DecodeString(tokenArr[0])
	if err != nil {
		return nil, err
	}

	type Header struct {
		Alg string   `json:"alg"`
		X5c []string `json:"x5c"`
	}
	var header Header
	err = json.Unmarshal(headerByte, &header)
	if err != nil {
		return nil, err
	}

	certByte, err := base64.StdEncoding.DecodeString(header.X5c[index])
	if err != nil {
		return nil, err
	}

	return certByte, nil
}

// openssl x509 -inform der -in AppleRootCA-G3.cer -out apple_root.pem
const rootPEM = `
-----BEGIN CERTIFICATE-----
MIICQzCCAcmgAwIBAgIILcX8iNLFS5UwCgYIKoZIzj0EAwMwZzEbMBkGA1UEAwwS
QXBwbGUgUm9vdCBDQSAtIEczMSYwJAYDVQQLDB1BcHBsZSBDZXJ0aWZpY2F0aW9u
IEF1dGhvcml0eTETMBEGA1UECgwKQXBwbGUgSW5jLjELMAkGA1UEBhMCVVMwHhcN
MTQwNDMwMTgxOTA2WhcNMzkwNDMwMTgxOTA2WjBnMRswGQYDVQQDDBJBcHBsZSBS
b290IENBIC0gRzMxJjAkBgNVBAsMHUFwcGxlIENlcnRpZmljYXRpb24gQXV0aG9y
aXR5MRMwEQYDVQQKDApBcHBsZSBJbmMuMQswCQYDVQQGEwJVUzB2MBAGByqGSM49
AgEGBSuBBAAiA2IABJjpLz1AcqTtkyJygRMc3RCV8cWjTnHcFBbZDuWmBSp3ZHtf
TjjTuxxEtX/1H7YyYl3J6YRbTzBPEVoA/VhYDKX1DyxNB0cTddqXl5dvMVztK517
IDvYuVTZXpmkOlEKMaNCMEAwHQYDVR0OBBYEFLuw3qFYM4iapIqZ3r6966/ayySr
MA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgEGMAoGCCqGSM49BAMDA2gA
MGUCMQCD6cHEFl4aXTQY2e3v9GwOAEZLuN+yRhHFD/3meoyhpmvOwgPUnPWTxnS4
at+qIxUCMG1mihDK1A3UT82NQz60imOlM27jbdoXt2QfyFMm+YhidDkLF1vLUagM
6BgD56KyKA==
-----END CERTIFICATE-----
`

func (a *AppStore) verifyCert(certByte, intermediaCertStr []byte) error {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return errors.New("failed to parse root certificate")
	}

	interCert, err := x509.ParseCertificate(intermediaCertStr)
	if err != nil {
		return errors.New("failed to parse intermedia certificate")
	}
	intermedia := x509.NewCertPool()
	intermedia.AddCert(interCert)

	cert, err := x509.ParseCertificate(certByte)
	if err != nil {
		return err
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermedia,
	}

	chains, err := cert.Verify(opts)
	if err != nil {
		return err
	}

	for _, ch := range chains {
		for _, c := range ch {
			fmt.Printf("%+v, %s, %+v \n", c.AuthorityKeyId, c.Subject.Organization, c.ExtKeyUsage)
		}
	}

	return nil
}

func (a *AppStore) ExtractClaims(tokenStr string) (*JWSTransaction, error) {
	tran := &JWSTransaction{}

	rootCertStr, err := a.extractHeaderByIndex(tokenStr, 2)
	if err != nil {
		return nil, err
	}
	intermediaCertStr, err := a.extractHeaderByIndex(tokenStr, 1)
	if err != nil {
		return nil, err
	}
	if err = a.verifyCert(rootCertStr, intermediaCertStr); err != nil {
		return nil, err
	}

	_, err = jwt.ParseWithClaims(tokenStr, tran, func(token *jwt.Token) (interface{}, error) {
		return a.extractPublicKeyFromToken(tokenStr)
	})
	if err != nil {
		return nil, err
	}

	return tran, nil
}
