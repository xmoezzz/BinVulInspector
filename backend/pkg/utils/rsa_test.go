package utils_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"bin-vul-inspector/pkg/utils"
)

func TestRsaEncrypt(t *testing.T) {
	var err error
	publicKey := []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsuv0oFMzjoA8+fPEy5h6
CjyX+C4KgZ5+sDoI59TiZ6sXlAvG3wq+e8ULvjT66q+QiqiGHJbASmyAaepbemh0
6n/KPK9xaBE54fuv89mpRHX2WwDx//aFF1dV71nryzG0KY3yZljRR4p1UEPfNngk
ehtjNMbnGNCpSttuNGydN9BeKpHelii3kIa3HCUvYnBfHKDgkghSVjWwVZs203Di
lclSW0NBMCbO3L2CWQ/rtqeRCnqUSYxl7xWqbM+HhPOb9H+nbNk4jiWf1IVCXFJN
PuxpEbNVFbO/CqoB/mLdpeCu6HS2BstWNb2g3/nWymyE6HLf6HEgdfIklTl41D01
7wIDAQAB
-----END PUBLIC KEY-----`)
	{
		plainText := []byte("V6oKxpZMvcSE1xhlplM961b0ZGYDbVIN")
		var data []byte
		if data, err = utils.RsaEncrypt(plainText, publicKey); err != nil {
			t.Fatal(err)
		}
		t.Logf("base64 encrypted: %s", base64.StdEncoding.EncodeToString(data))
	}
}

func TestRsaDecrypt(t *testing.T) {
	var err error
	privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAsuv0oFMzjoA8+fPEy5h6CjyX+C4KgZ5+sDoI59TiZ6sXlAvG
3wq+e8ULvjT66q+QiqiGHJbASmyAaepbemh06n/KPK9xaBE54fuv89mpRHX2WwDx
//aFF1dV71nryzG0KY3yZljRR4p1UEPfNngkehtjNMbnGNCpSttuNGydN9BeKpHe
lii3kIa3HCUvYnBfHKDgkghSVjWwVZs203DilclSW0NBMCbO3L2CWQ/rtqeRCnqU
SYxl7xWqbM+HhPOb9H+nbNk4jiWf1IVCXFJNPuxpEbNVFbO/CqoB/mLdpeCu6HS2
BstWNb2g3/nWymyE6HLf6HEgdfIklTl41D017wIDAQABAoIBAQCDiRwKUaBxfq1V
RHTFCI+PvwQqHA71Q8P3YnLxnHvlos5utEm753YqH66GYwSkS/WDOml90wYCsMmn
E/e0gd6SFuhivMgurZtUG2g7aSUbg21dcdB3UJB/nGE82WqTszKz6fruaxVP9uZP
39XVgXWvnzrLrf5vK9eJhM/8Em1yfaywtZZdHYU78iQfvNLDCDTA53nYR5ShERVg
cdqmb2BUT6iXlC265qrpdntoT8UGf1Gp0o/VSkQKdhwHipmfbWTUV+Fq9G7uWOBO
lVNwjm1ngY9IYI2as2VV68jM6miCBgozMdhVPs/Z3Z/m/VZwP3oTezL8OZtFXD45
Oxqr9vaBAoGBAOzYlqpKFov69CkmES7+XYWOy/qyFeJyMvKVHJ/2gcF/gOBr1F1k
vFeCfq+/RuEYcAzY0s9cBwbsDfvkL1DQ568Xqm1qLgXWkj3q9RZVrsGAc0kqZFpS
7B8+avNOu/hzkpXatkAYN28yYKPnymimkTEh4p0cq8HpICxV6z9ehdLxAoGBAMFk
KYrI+qwpNBpfvwQnUKKg9Wwy9MUoJAhdhTmqEDpLpqLoX3PmLA5Q/h85+nL2GUr5
HpbqG/ehW68kKs1sXQGpI+tgz//XpgH/YckqUFPWMpyLujxwRJL5kwmJ/qTcBeXp
MAAFiUUOjlnYNpXKeZyo7d68qBAxhGVrW3dKpdbfAoGAYgj4vD30fTaAD/RA0pnZ
LipARmGma1fnvL953MCVTvmu57Xabln/F53dQHPFK/EImFi7Ubd+9R+KXkRCTYpb
C/+YvLdhm2sIl3aEwhzvPAsmLRfN+BEwyXH1pQZnCd0UxNCF9ZvQfkd09wM/pfek
S5kCCxROB/KuLYvW1yER9ZECgYANf04e75P/PAj05kXQpmXMU+uNF6lZsUmCg/Ru
Z94mE22X5Rv0XNYqUaDK0SMXrvFo+CYYZlJ5X/ukJ6QNHkkHeqSVIvahZo2hig9r
GNbuYv65Sk8/NJ60m1KV0dnB69FFkJbXCYvhE/j/cEWvAqimNGwVpZkdODeDVJDX
rJAShQKBgQCPOnqLKuk8IKleuhBvoy21B79uoaAhbaavWRFYahbSj433WbxNqXL2
Jq88PBwxFNJaqDHJdp+oZ4JqbcQUQBvj46oObs+/V/ZV6VYGkH2ANNy+u/JpZJ46
R22NuX6J1ghFBA191ghJJZ9XAXNC0IwwZXJ4/hQHyg7TsHGANkGuiA==
-----END RSA PRIVATE KEY-----`)
	{
		var data []byte
		encrypted := "MZzklltmRr6sPxT7sHxjpmjivpGr5sJDBd6ZTTayeVmAERSbZBfW4ItZt5K7gV9aRUhKgMSyspX8zGjunNNafXAf288urNKB40hIYeChQN1K1DC5zWaGnxF5yIu4/Riu5R9rS4qtf4kSfUEIlvnZttIaCgKfQpD/ePrX5jc1aeWqt2tqzJqcXeCxr3LJaLSRG6uHhLWgVhDVoKoGjGwqvPis5NG9tMuXoFRWSt09DWw4Sw6gIXL8L8BxwVxI1TEqgqPSk/xql/aJlQbajJBZRLvNNfsYS+Xb5L4tE/M+iovU94JZ5pnbH4hK/Z+0XdcJ3Bltjg7+hE3J77Pg1CbvvQ=="
		if data, err = base64.StdEncoding.DecodeString(encrypted); err != nil {
			t.Fatal(err)
		}
		if data, err = utils.RsaDecrypt(data, privateKey); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "V6oKxpZMvcSE1xhlplM961b0ZGYDbVIN", string(data))
	}
}
