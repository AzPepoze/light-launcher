package main

import "testing"

func TestCookieHeaderMatchesDomainPathAndSecure(t *testing.T) {
	jar := newCookieJar()
	jar.add(cookie{Name: "sid", Value: "s", Domain: ".example.com", Path: "/"})
	jar.add(cookie{Name: "token", Value: "t", Domain: "sub.example.com", Path: "/download"})
	jar.add(cookie{Name: "wide", Value: "w", Domain: "example.com", Path: "/"})
	jar.add(cookie{Name: "secure", Value: "sec", Domain: "sub.example.com", Path: "/", Secure: true})
	jar.add(cookie{Name: "other", Value: "o", Domain: "notexample.com", Path: "/"})
	jar.add(cookie{Name: "path", Value: "p", Domain: "sub.example.com", Path: "/downloads"})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "https sends every match with longer paths first",
			url:  "https://sub.example.com/download/file?token=1",
			want: "token=t; sid=s; wide=w; secure=sec",
		},
		{
			name: "http drops secure cookies",
			url:  "http://sub.example.com/download/file",
			want: "token=t; sid=s; wide=w",
		},
		{
			name: "path must end at a segment boundary",
			url:  "https://sub.example.com/downloads/file",
			want: "path=p; sid=s; wide=w; secure=sec",
		},
		{
			name: "other host matches nothing",
			url:  "https://example.org/download/file",
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := jar.header(test.url); got != test.want {
				t.Fatalf("header = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCookieJarUpdatesAndRemovesByIdentifier(t *testing.T) {
	jar := newCookieJar()
	jar.add(cookie{Name: "sid", Value: "old", Domain: "example.com", Path: "/"})
	jar.add(cookie{Name: "sid", Value: "new", Domain: ".example.com", Path: ""})

	if got, want := jar.header("https://example.com/"), "sid=new"; got != want {
		t.Fatalf("header after update = %q, want %q", got, want)
	}

	jar.remove(cookie{Name: "sid", Domain: ".EXAMPLE.com", Path: "/"})
	if got := jar.header("https://example.com/"); got != "" {
		t.Fatalf("header after removal = %q, want empty", got)
	}
}

func TestCookieJarIgnoresCookiesWithoutName(t *testing.T) {
	jar := newCookieJar()
	jar.add(cookie{Value: "orphan", Domain: "example.com", Path: "/"})
	if got := jar.header("https://example.com/"); got != "" {
		t.Fatalf("header = %q, want empty", got)
	}
}

func TestCookieHeaderRejectsNonWebURL(t *testing.T) {
	jar := newCookieJar()
	jar.add(cookie{Name: "sid", Value: "s", Domain: "example.com", Path: "/"})
	for _, raw := range []string{"", "blob:https://example.com/file", "not a url", "https:///file"} {
		if got := jar.header(raw); got != "" {
			t.Fatalf("header(%q) = %q, want empty", raw, got)
		}
	}
}

func TestDomainMatches(t *testing.T) {
	tests := []struct {
		domain string
		host   string
		want   bool
	}{
		{domain: ".example.com", host: "example.com", want: true},
		{domain: "example.com", host: "sub.example.com", want: true},
		{domain: ".EXAMPLE.com.", host: "Sub.Example.Com", want: true},
		{domain: "example.com", host: "notexample.com", want: false},
		{domain: "example.com", host: "example.org", want: false},
		{domain: "", host: "example.com", want: false},
		{domain: "127.0.0.1", host: "127.0.0.1", want: true},
		{domain: "127.0.0.1", host: "sub.127.0.0.1", want: false},
	}
	for _, test := range tests {
		if got := domainMatches(test.domain, test.host); got != test.want {
			t.Errorf("domainMatches(%q, %q) = %v, want %v", test.domain, test.host, got, test.want)
		}
	}
}

func TestPathMatches(t *testing.T) {
	tests := []struct {
		cookiePath  string
		requestPath string
		want        bool
	}{
		{cookiePath: "/", requestPath: "/", want: true},
		{cookiePath: "", requestPath: "/", want: true},
		{cookiePath: "/download", requestPath: "/download", want: true},
		{cookiePath: "/download", requestPath: "/download/file", want: true},
		{cookiePath: "/download", requestPath: "/downloads/file", want: false},
		{cookiePath: "/download/", requestPath: "/download/file", want: true},
		{cookiePath: "/a", requestPath: "/b", want: false},
	}
	for _, test := range tests {
		if got := pathMatches(test.cookiePath, test.requestPath); got != test.want {
			t.Errorf("pathMatches(%q, %q) = %v, want %v", test.cookiePath, test.requestPath, got, test.want)
		}
	}
}
