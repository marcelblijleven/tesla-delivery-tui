package model

import (
	"testing"
)

func TestDecodeOptionCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		// Autopilot
		{"Autopilot Basic", "APBS", "Autopilot - Basic"},
		{"FSD Capability", "APF2", "Full Self-Driving Capability"},
		{"HW3", "APH3", "Autopilot 3.0 Hardware (HW3)"},

		// Paint Colors
		{"Pearl White", "PPSW", "Pearl White Multi-Coat"},
		{"Solid Black", "PBSB", "Solid Black"},
		{"Red Multi-Coat", "PPMR", "Red Multi-Coat"},
		{"Midnight Silver", "PMTG", "Midnight Silver Metallic"},

		// Interior
		{"Black Premium Interior", "IPB1", "Black Premium Interior"},
		{"White Premium Interior", "IPW1", "White Premium Interior"},
		{"All Black Interior", "IBB1", "All Black Interior"},

		// Wheels
		{"19 Gemini Wheels", "WY19B", "19\" Gemini Wheels"},
		{"20 Induction Wheels", "WY20P", "20\" Induction Wheels"},
		{"18 Aero Wheels", "W38B", "18\" Aero Wheels"},

		// Model specific
		{"Model Y Long Range AWD", "MTY52", "Model Y Long Range AWD"},
		{"Model 3 Long Range AWD", "MT307", "Model 3 Long Range AWD"},

		// Other
		{"Tow Hitch", "TW01", "Tow Hitch"},
		{"No Tow Hitch", "TW00", "No Tow Hitch"},
		{"Pay Per Use Supercharging", "SC04", "Pay Per Use Supercharging"},

		// Unknown code
		{"Unknown code", "XXXXX", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecodeOptionCode(tt.code); got != tt.want {
				t.Errorf("DecodeOptionCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestDecodeOptions(t *testing.T) {
	tests := []struct {
		name       string
		optionsStr string
		wantLen    int
		wantFirst  DecodedOption
	}{
		{
			name:       "single option",
			optionsStr: "PPSW",
			wantLen:    1,
			wantFirst:  DecodedOption{Code: "PPSW", Description: "Pearl White Multi-Coat"},
		},
		{
			name:       "multiple options",
			optionsStr: "PPSW,IPB1,WY19B",
			wantLen:    3,
			wantFirst:  DecodedOption{Code: "PPSW", Description: "Pearl White Multi-Coat"},
		},
		{
			name:       "options with spaces",
			optionsStr: "PPSW, IPB1, WY19B",
			wantLen:    3,
			wantFirst:  DecodedOption{Code: "PPSW", Description: "Pearl White Multi-Coat"},
		},
		{
			name:       "empty string",
			optionsStr: "",
			wantLen:    0,
		},
		{
			name:       "unknown options included",
			optionsStr: "PPSW,XXXXX,IPB1",
			wantLen:    3,
			wantFirst:  DecodedOption{Code: "PPSW", Description: "Pearl White Multi-Coat"},
		},
		{
			name:       "empty entries filtered",
			optionsStr: "PPSW,,IPB1",
			wantLen:    2,
			wantFirst:  DecodedOption{Code: "PPSW", Description: "Pearl White Multi-Coat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeOptions(tt.optionsStr)

			if len(got) != tt.wantLen {
				t.Errorf("DecodeOptions() returned %d options, want %d", len(got), tt.wantLen)
				return
			}

			if tt.wantLen > 0 {
				if got[0].Code != tt.wantFirst.Code {
					t.Errorf("First option code = %q, want %q", got[0].Code, tt.wantFirst.Code)
				}
				if got[0].Description != tt.wantFirst.Description {
					t.Errorf("First option description = %q, want %q", got[0].Description, tt.wantFirst.Description)
				}
			}
		})
	}
}

func TestCategorizeOptions(t *testing.T) {
	options := DecodeOptions("MDLY,PPSW,IPB1,WY19B,APBS,SC04,TW01")
	categories := CategorizeOptions(options)

	tests := []struct {
		category string
		wantLen  int
	}{
		{"Model", 1},     // MDLY
		{"Paint", 1},     // PPSW
		{"Interior", 1},  // IPB1
		{"Wheels", 1},    // WY19B
		{"Autopilot", 1}, // APBS
		{"Charging", 1},  // SC04
		{"Other", 1},     // TW01
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := len(categories[tt.category])
			if got != tt.wantLen {
				t.Errorf("Category %q has %d options, want %d", tt.category, got, tt.wantLen)
			}
		})
	}
}

func TestCategorizeOptions_PaintPrefixes(t *testing.T) {
	// Test all paint prefixes are properly categorized
	paintCodes := []string{"PPSW", "PMTG", "PBSB", "PN00", "PR00"}
	options := make([]DecodedOption, len(paintCodes))
	for i, code := range paintCodes {
		options[i] = DecodedOption{Code: code, Description: DecodeOptionCode(code)}
	}

	categories := CategorizeOptions(options)
	if len(categories["Paint"]) != len(paintCodes) {
		t.Errorf("Paint category has %d options, want %d", len(categories["Paint"]), len(paintCodes))
	}
}

func TestCategorizeOptions_EmptyInput(t *testing.T) {
	categories := CategorizeOptions(nil)

	// All categories should exist but be empty
	expectedCategories := []string{"Model", "Paint", "Interior", "Wheels", "Autopilot", "Charging", "Other"}
	for _, cat := range expectedCategories {
		if _, ok := categories[cat]; !ok {
			t.Errorf("Category %q missing from result", cat)
		}
		if len(categories[cat]) != 0 {
			t.Errorf("Category %q should be empty, has %d items", cat, len(categories[cat]))
		}
	}
}

func TestDecodeOptions_RealWorldExample(t *testing.T) {
	// Real option string from demo data
	optionsStr := "APBS,IPB11,PPSW,SC04,MDLY,WY19P,MTY52,STY5S,CPF0,TW01"
	options := DecodeOptions(optionsStr)

	if len(options) != 10 {
		t.Errorf("Expected 10 options, got %d", len(options))
	}

	// Check specific options
	expectedCodes := map[string]bool{
		"APBS":  true,
		"IPB11": true,
		"PPSW":  true,
		"SC04":  true,
		"MDLY":  true,
		"WY19P": true,
		"MTY52": true,
		"STY5S": true,
		"CPF0":  true,
		"TW01":  true,
	}

	for _, opt := range options {
		if !expectedCodes[opt.Code] {
			t.Errorf("Unexpected option code: %s", opt.Code)
		}
	}
}
