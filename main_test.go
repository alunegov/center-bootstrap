package main

import (
	"strings"
	"testing"
)

func TestExtractSerials(t *testing.T) {
	got := extractSerials([]byte(`[gsm.serial]: [VPR081831768                                                10P]
		[gsm.serial]: [VPR081831768]
		[serial]: [VPR081831768]
		[gsm.serial]: []
		[gsm.serial]: 
		[]: [VPR081831768]

	`))
	if len(got) != 3 {
		t.Errorf("length of extractSerials(ref) = %d, expected %d", len(got), 3)
	}
	for i, g := range got {
		if g.Value != "VPR081831768" {
			t.Errorf("extractSerials(ref)[%d] = %s, expected %s", i, g, "VPR081831768")
		}
	}

	got = extractSerials([]byte(``))
	if len(got) != 0 {
		t.Errorf("length of extractSerials('') = %d, expected %d", len(got), 0)
	}
}

func TestExtractAndroidID(t *testing.T) {
	cases := []struct {
		test string
		exp  string
	}{
		{`CenterId: id2 = 1`, "1"},
		{`CenterId : id2 = 2`, "2"},
		{` CenterId: id2 = 3`, "3"},
		{` CenterId : id2 = 4`, "4"},
		{` CenterId : id2 = 5 `, "5 "},
		{` CenterId : id2 =  6 `, " 6 "},
		{``, ""},
	}

	for i, c := range cases {
		got := extractAndroidID([]byte(c.test))
		if got != c.exp {
			t.Errorf("extractAndroidID(cases[%d]) = %s, expected %s", i, got, c.exp)
		}
	}
}

func TestExtractVersion(t *testing.T) {
	cases := []struct {
		test string
		exp  string
	}{
		{`versionName '1'`, "1"},
		{`versionName  '2'`, "2"},
		{` versionName '3'`, "3"},
		{` versionName  '4'`, "4"},
		{` versionName  '5' `, "5"},
		{`versionName "7"`, "7"},
		{`versionName  "8"`, "8"},
		{` versionName "9"`, "9"},
		{` versionName  "10"`, "10"},
		{` versionName  "11" `, "11"},
		{`versionName 13`, ""},
		{``, ""},
	}

	for i, c := range cases {
		got := extractVersion(strings.NewReader(c.test))
		if got != c.exp {
			t.Errorf("extractVersion(cases[%d]) = %s, expected %s", i, got, c.exp)
		}
	}
}
