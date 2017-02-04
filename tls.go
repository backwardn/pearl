package pearl

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/mmcloughlin/pearl/torkeys"
)

type ConnectionCertificates interface {
}

func GenerateConnectionCertificates(idKey torkeys.PrivateKey) (ConnectionCertificates, error) {
	// Reference: https://github.com/torproject/torspec/blob/master/tor-spec.txt#L225-L245
	//
	//	   In "in-protocol" (a.k.a. "the v3 handshake"), the initiator sends no
	//	   certificates, and the
	//	   responder sends a single connection certificate.  The choice of
	//	   ciphersuites must be as in a "renegotiation" handshake.  There are
	//	   additionally a set of constraints on the connection certificate,
	//	   which the initiator can use to learn that the in-protocol handshake
	//	   is in use.  Specifically, at least one of these properties must be
	//	   true of the certificate:
	//	      * The certificate is self-signed
	//	      * Some component other than "commonName" is set in the subject or
	//	        issuer DN of the certificate.
	//	      * The commonName of the subject or issuer of the certificate ends
	//	        with a suffix other than ".net".
	//	      * The certificate's public key modulus is longer than 1024 bits.
	//	   The initiator then sends a VERSIONS cell to the responder, which then
	//	   replies with a VERSIONS cell; they have then negotiated a Tor
	//	   protocol version.  Assuming that the version they negotiate is 3 or higher
	//	   (the only ones specified for use with this handshake right now), the
	//	   responder sends a CERTS cell, an AUTH_CHALLENGE cell, and a NETINFO
	//	   cell to the initiator, which may send either CERTS, AUTHENTICATE,
	//	   NETINFO if it wants to authenticate, or just NETINFO if it does not.
	//
	// Reference: https://github.com/torproject/tor/blob/master/src/common/tortls.c#L1061-L1066
	//
	//	  nickname = crypto_random_hostname(8, 20, "www.", ".net");
	//	#ifdef DISABLE_V3_LINKPROTO_SERVERSIDE
	//	  nn2 = crypto_random_hostname(8, 20, "www.", ".net");
	//	#else
	//	  nn2 = crypto_random_hostname(8, 20, "www.", ".com");
	//	#endif
	//

	linkCN := randomHostname(8, 20, "www.", ".net")
	idCN := randomHostname(8, 20, "www.", ".com")

	// Certificate lifetime is either set by the SSLKeyLifetime option or
	// generated to a reasonable looking value.
	//
	// BUG(mmcloughlin): SSLKeyLifetime option ignored when generating
	// certificates.
	lifetime := generateCertificateLifetime()

	fmt.Println(linkCN, idCN, lifetime)

	// Reference: https://github.com/torproject/tor/blob/master/src/common/tortls.c#L1080-L1082
	//
	//	    /* Create a link certificate signed by identity key. */
	//	    cert = tor_tls_create_certificate(rsa, identity, nickname, nn2,
	//	                                      key_lifetime);
	//

	//linkCert, err := openssl.NewCertificate(&openssl.CertificateInfo{
	//	CommonName: linkCN,
	//	Serial:     rand.Int63(),
	//	Issued:     issued,
	//	Expires:    expires,
	//}, tmpPk)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
}

// randomHostname generates a hostname starting with prefix, ending with
// suffix, and of length between min and max (inclusive).
//
// Reference: https://github.com/torproject/tor/blob/master/src/common/crypto.c#L3172-L3181
//
//	/** Generate and return a new random hostname starting with <b>prefix</b>,
//	 * ending with <b>suffix</b>, and containing no fewer than
//	 * <b>min_rand_len</b> and no more than <b>max_rand_len</b> random base32
//	 * characters. Does not check for failure.
//	 *
//	 * Clip <b>max_rand_len</b> to MAX_DNS_LABEL_SIZE.
//	 **/
//	char *
//	crypto_random_hostname(int min_rand_len, int max_rand_len, const char *prefix,
//	                       const char *suffix)
//
func randomHostname(min, max int, prefix, suffix string) string {
	// Reference: https://github.com/torproject/tor/blob/master/src/common/util_format.h#L23-L25
	//
	//	/** Characters that can appear (case-insensitively) in a base32 encoding. */
	//	#define BASE32_CHARS "abcdefghijklmnopqrstuvwxyz234567"
	//	void base32_encode(char *dest, size_t destlen, const char *src, size_t srclen);
	//
	alphabet := "abcdefghijklmnopqrstuvwxyz234567"
	n := min + rand.Intn(max-min+1)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return prefix + string(b) + suffix
}

// generateCertificateLifetime generates a reasonable looking certificate
// lifetime.
//
// Reference: https://github.com/torproject/tor/blob/master/src/or/router.c#L702-L717
//
//	  if (!lifetime) { /* we should guess a good ssl cert lifetime */
//
//	    /* choose between 5 and 365 days, and round to the day */
//	    unsigned int five_days = 5*24*3600;
//	    unsigned int one_year = 365*24*3600;
//	    lifetime = crypto_rand_int_range(five_days, one_year);
//	    lifetime -= lifetime % (24*3600);
//
//	    if (crypto_rand_int(2)) {
//	      /* Half the time we expire at midnight, and half the time we expire
//	       * one second before midnight. (Some CAs wobble their expiry times a
//	       * bit in practice, perhaps to reduce collision attacks; see ticket
//	       * 8443 for details about observed certs in the wild.) */
//	      lifetime--;
//	    }
//	  }
//
func generateCertificateLifetime() time.Duration {
	days := 5 + rand.Intn(360)
	wobble := rand.Intn(2)
	return time.Duration(days*24)*time.Hour - time.Duration(wobble)*time.Second
}

// generateCertificateSerial generates a serial number for a certificate. This
// copies the convention of openssl and returns a 64-bit integer. Returns
// big.Int so it can be used with
// https://godoc.org/github.com/mmcloughlin/openssl#CertificateInfo.
//
// Reference: https://github.com/torproject/tor/blob/master/src/common/tortls.c#L468-L470
//
//	  /* OpenSSL generates self-signed certificates with random 64-bit serial
//	   * numbers, so let's do that too. */
//	#define SERIAL_NUMBER_SIZE 8
//
// Reference: https://github.com/torproject/tor/blob/master/src/common/tortls.c#L502-L508
//
//	  { /* our serial number is 8 random bytes. */
//	    crypto_rand((char *)serial_tmp, sizeof(serial_tmp));
//	    if (!(serial_number = BN_bin2bn(serial_tmp, sizeof(serial_tmp), NULL)))
//	      goto error;
//	    if (!(BN_to_ASN1_INTEGER(serial_number, X509_get_serialNumber(x509))))
//	      goto error;
//	  }
//
func generateCertificateSerial() (*big.Int, error) {
	serialBytes := make([]byte, 8)
	_, err := cryptorand.Read(serialBytes)
	if err != nil {
		return nil, err
	}
	serial := big.NewInt(0)
	return serial.SetBytes(serialBytes), nil
}