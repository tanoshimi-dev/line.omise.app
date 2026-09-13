package service

import "testing"

func TestValidateStatus(t *testing.T) {
	cases := []struct {
		status  string
		wantErr bool
	}{
		{"draft", false},
		{"published", false},
		{"archived", true},
		{"", true},
	}
	for _, c := range cases {
		err := ValidateStatus(c.status)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateStatus(%q) error = %v, wantErr %v", c.status, err, c.wantErr)
		}
	}
}

func TestValidateArticleCategory(t *testing.T) {
	cases := []struct {
		category string
		wantErr  bool
	}{
		{"line-operation", false},
		{"ai", false},
		{"line-setup", true},
		{"", true},
	}
	for _, c := range cases {
		err := ValidateArticleCategory(c.category)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateArticleCategory(%q) error = %v, wantErr %v", c.category, err, c.wantErr)
		}
	}
}

func TestValidateRelatedDemoApp(t *testing.T) {
	cases := []struct {
		app     string
		wantErr bool
	}{
		{"", false}, // optional association
		{"membership", false},
		{"salon-reservation", false},
		{"sweets-shop", false},
		{"bogus-app", true},
	}
	for _, c := range cases {
		err := ValidateRelatedDemoApp(c.app)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateRelatedDemoApp(%q) error = %v, wantErr %v", c.app, err, c.wantErr)
		}
	}
}
