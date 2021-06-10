package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// CreateSign returns signature for given u
func CreateSign(key string, u *url.URL) string {
	q := u.Query()
	q.Del("s")

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(q.Encode()))
	sign := mac.Sum(nil)

	return hex.EncodeToString(sign)
}

// CreateSign returns signature for given u
func createCompatSign(key string, u *url.URL) string {
	canonicalRawQuery := canonicalizeQuery(u.RawQuery)

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(canonicalRawQuery))
	sign := mac.Sum(nil)

	return hex.EncodeToString(sign)
}

// SignURL adds signature to u using key
func SignURL(key string, u *url.URL) {
	q := u.Query()
	q.Del("s")
	q.Add("s", CreateSign(key, u))
	u.RawQuery = q.Encode()
}

// ValidSign validates signature of u using key
func ValidSign(key string, u *url.URL) bool {
	return createCompatSign(key, u) == u.Query().Get("s")
}

func canonicalizeQuery(query string) string {
	kvp := make(map[string][]string)

	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}

		// ignore signature
		if key == "s" {
			continue
		}

		kvp[key] = append(kvp[key], value)
	}

	var buf bytes.Buffer
	keys := make([]string, 0, len(kvp))
	for k := range kvp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := kvp[k]
		prefix := k + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}

			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}

	return buf.String()
}
