package appleTools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	gopkcs12 "software.sslmate.com/src/go-pkcs12"
)

type Cert struct {
	CertID      string `json:"cert_id"`
	CertContent string `json:"-" gorm:"type:text"`
	P12Content  string `json:"-" gorm:"type:text"`
	P12Password string `json:"p12_password"`
}

func (c *Cert) CreateCsr(email string) (csr string, pKey *rsa.PrivateKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	emailAddress := email
	subj := pkix.Name{
		CommonName:         "test.com",
		Country:            []string{"AU"},
		Province:           []string{"Some-State"},
		Locality:           []string{"MyCity"},
		Organization:       []string{"Company Ltd1"},
		OrganizationalUnit: []string{"IT"},
	}
	rawSubj := subj.ToRDNSequence()
	rawSubj = append(rawSubj, []pkix.AttributeTypeAndValue{
		{Type: asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}, Value: emailAddress},
	})

	asn1Subj, _ := asn1.Marshal(rawSubj)
	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		EmailAddresses:     []string{emailAddress},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	return base64.StdEncoding.EncodeToString(csrBytes), privateKey
}
func (c *Cert) ToP12(priKey *rsa.PrivateKey, password string) (err error) {

	var certBuf []byte

	certBuf, err = base64.StdEncoding.DecodeString(c.CertContent)
	if err != nil {
		return err
	}
	crt, err := x509.ParseCertificate(certBuf)
	if err != nil {
		return errors.New("证书解析异常 :" + err.Error())
	}
	pfx, err := gopkcs12.Encode(rand.Reader, priKey, crt, nil, password)
	if err != nil {
		return errors.New("证书转换异常：" + err.Error())
	}
	c.P12Content = string(pfx)
	c.P12Password = password
	return nil
}
