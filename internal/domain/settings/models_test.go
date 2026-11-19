package settings

import "testing"

func TestValidateRejectsNonHTTPBackgroundURLs(t *testing.T) {
	for _, u := range []string{"javascript:alert(1)", "data:text/html,x", "ftp://host/a.jpg", "not a url"} {
		req := &UpdateSettingsRequest{BackgroundImageURL: &u}
		if err := req.Validate(); err == nil {
			t.Errorf("%q accepted", u)
		}
	}
	for _, u := range []string{"", "https://example.com/a.jpg", "http://localhost:9000/b.png"} {
		req := &UpdateSettingsRequest{BackgroundImageURL: &u}
		if err := req.Validate(); err != nil {
			t.Errorf("%q rejected: %v", u, err)
		}
	}
}

func TestValidateRejectsUnknownFonts(t *testing.T) {
	bad := "Comic Sans'; } body { display:none"
	if err := (&UpdateSettingsRequest{FontFamily: &bad}).Validate(); err == nil {
		t.Error("unknown font accepted")
	}
	good := string(FontFamilyOpenSans)
	if err := (&UpdateSettingsRequest{FontFamily: &good}).Validate(); err != nil {
		t.Errorf("%q rejected: %v", good, err)
	}
}
