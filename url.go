// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import (
	"regexp"
	"strings"
)

// A URL is an immutable string-like type that is safe to use in URL contexts in
// DOM APIs and HTML documents.
//
// URL guarantees that its value as a string will not cause untrusted script execution
// when evaluated as a hyperlink URL in a browser.
//
// Values of this type are guaranteed to be safe to use in URL/hyperlink contexts,
// such as assignment to URL-valued DOM properties, in the sense that the use
// will not result in a Cross-site Scripting (XSS) vulnerability. Similarly, URLs can
// be interpolated into the URL context of an HTML template (e.g. inside a href attribute).
// However, appropriate HTML-escaping must still be applied.
//
// Note that this type's contract does not imply any guarantees regarding the resource
// the URL refers to. In particular, URLs are not safe to use in a context
// where the referred-to resource is interpreted as trusted code, e.g., as the src of
// a script tag. For safely loading trusted resources, use the TrustedResourceURL type.
type URL struct {
	// We declare a URL not as a string but as a struct wrapping a string
	// to prevent construction of URL values through string conversion.
	str string
}

// InnocuousURL is an innocuous URL generated by URLSanitized when passed an unsafe URL.
//
// about:invalid is registered in http://www.w3.org/TR/css3-values/#about-invalid,
// and "references a non-existent document with a generic error condition. It can be
// used when a URI is necessary, but the default value shouldn't be resolveable as any
// type of document."
//
// http://tools.ietf.org/html/rfc6694#section-2.1 permits about URLs to contain
// a fragment, which is not to be considered when determining if an about URL is
// well-known.
const InnocuousURL = "about:invalid#zGoSafez"

// URLSanitized returns a URL whose value is url, validating that the input string matches
// a pattern of commonly used safe URLs. If url fails validation, this method returns a
// URL containing InnocuousURL.
//
// url may be a URL with an explicit scheme that is not javascript, or a relative URL,
// i.e. a URL without a scheme. Specifically, a relative URL may be scheme-relative,
// absolute-path-relative, or path-relative. See
// http://url.spec.whatwg.org/#concept-relative-url.
//
// If the URL has an explicit scheme, it should only contain characters allowed by the URL
// specification, i.e. alphanumeric or [+-.].
//
// No attempt is made at validating that the URL percent-decodes to structurally valid or
// interchange-valid UTF-8 since the percent-decoded representation is unsafe to use in an
// HTML context regardless of UTF-8 validity.
func URLSanitized(url string) URL {
	if !isSafeURL(url) {
		return URL{InnocuousURL}
	}
	return URL{url}
}

// safeURLPattern matches URLs that
//
//		(a) Start with (capture) an explicit scheme. The scheme needs to be further validated to ban
//	     javascript; or
//		(b) Contain no scheme. To ensure that the URL cannot be interpreted as a
//		    disallowed scheme URL, ':' may only appear after one of the runes [/?#].
//
// Warning: Matching this regex is not enough to validate the SafeUrl contract. See the isSafeUrl
// implementation for more information.
//
// The origin (RFC 6454) in which a URL is loaded depends on
// its scheme.  We assume that the scheme used by the current document is HTTPS, HTTP, or
// something equivalent.  We allow relative URLs unless in a particularly sensitive context
// called a "TrustedResourceUrl" context. In a non-TrustedResourceURL context we allow absolute
// URLs whose scheme is on a white-list.
//
// The position of the first colon (':') character determines whether a URL is absolute or relative.
// Looking at the prefix leading up to the first colon allows us to identify relative and absolute URLs,
// extract the scheme, and minimize the risk of a user-agent concluding a URL specifies a scheme not in
// our allowlist.
//
// According to RFC 3986 Section 3, the normative interpretation of the canonicial WHATWG specification
// (https://url.spec.whatwg.org/#url-scheme-string), colons can appear in a URL in these locations:
//   - A colon after a non-empty run of (ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )) ends a scheme.
//     If the colon after the scheme is not followed by "//" then any subsequent colons are part
//     of an opaque URI body.
//   - Otherwise, a colon after a hash (#) must be in the fragment.
//   - Otherwise, a colon after a (?) must be in the query.
//   - Otherwise, a colon after a single solidus ("/") must be in the path.
//   - Otherwise, a colon after a double solidus ("//") must be in the authority (before port).
//   - Otherwise, a colon after a valid protocol must be in the opaque part of the URL.
var safeURLPattern = regexp.MustCompile(`^(?:([a-z0-9+.-]+):|[^&:\/?#]*(?:[\/?#]|$))`)

// isSafeURL matches url to a subset of URLs that will not cause script execution if used in
// a URL context within a HTML document. Specifically, this method returns true if url:
//
//	(a) Starts with an explicit scheme that is not javascript; or
//	(b) Contains no scheme. To ensure that the URL cannot be interpreted as a
//	    disallowed scheme URL, the runes ':', and '&' may only appear
//	    after one of the runes [/?#].
func isSafeURL(url string) bool {
	// Ignore case.
	url = strings.ToLower(url)
	submatches := safeURLPattern.FindStringSubmatch(url)
	if submatches == nil {
		// No match
		return false
	}
	if len(submatches) == 0 {
		// Implicit URL scheme. This is safe
		return true
	}
	// Block javascript: URLs
	return len(submatches) == 2 && submatches[1] != "javascript"
}

// String returns the string form of the URL.
func (u URL) String() string {
	return u.str
}
