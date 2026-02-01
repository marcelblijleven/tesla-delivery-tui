package model

import (
	"testing"
)

func TestDecodeVIN(t *testing.T) {
	tests := []struct {
		name     string
		vin      string
		wantNil  bool
		expected *VINInfo
	}{
		{
			name:    "invalid VIN - too short",
			vin:     "5YJ3E1EA",
			wantNil: true,
		},
		{
			name:    "invalid VIN - too long",
			vin:     "5YJ3E1EA1LF1234567890",
			wantNil: true,
		},
		{
			name:    "invalid VIN - empty",
			vin:     "",
			wantNil: true,
		},
		{
			name:    "Model 3 from Fremont",
			vin:     "5YJ3AAEE1LF123456", // A=body LHD, E=fuel electric, E=Long Range AWD
			wantNil: false,
			expected: &VINInfo{
				VIN:                "5YJ3AAEE1LF123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Fremont, CA / Austin, TX, USA",
				Model:              "Model 3",
				BodyType:           "Sedan 4-door, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - Long Range, AWD",
				ModelYear:          "2020",
				ManufacturingPlant: "Fremont, CA, USA",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "Model Y from Berlin",
			vin:     "XP7YACEF9TB123456",
			wantNil: false,
			expected: &VINInfo{
				VIN:                "XP7YACEF9TB123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Berlin, Germany",
				Model:              "Model Y",
				BodyType:           "SUV 5-door, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - Long Range, AWD",
				ModelYear:          "2026",
				ManufacturingPlant: "Berlin, Germany",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "Model S from Fremont",
			vin:     "5YJSA1E2XLF123456",
			wantNil: false,
			expected: &VINInfo{
				VIN:                "5YJSA1E2XLF123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Fremont, CA / Austin, TX, USA",
				Model:              "Model S",
				BodyType:           "Hatchback 5-door, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - Standard",
				ModelYear:          "2020",
				ManufacturingPlant: "Fremont, CA, USA",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "Model X from Fremont",
			vin:     "5YJXCAE2XLF123456",
			wantNil: false,
			expected: &VINInfo{
				VIN:                "5YJXCAE2XLF123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Fremont, CA / Austin, TX, USA",
				Model:              "Model X",
				BodyType:           "Unknown", // 'C' not in Model X body types
				FuelType:           "Electric",
				Powertrain:         "Unknown", // Model X uses different powertrain codes
				ModelYear:          "2020",
				ManufacturingPlant: "Fremont, CA, USA",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "Model Y from Shanghai",
			vin:     "LRWYACEF5NC123456", // A=body LHD, E=fuel, F=LR AWD
			wantNil: false,
			expected: &VINInfo{
				VIN:                "LRWYACEF5NC123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Shanghai, China",
				Model:              "Model Y",
				BodyType:           "SUV 5-door, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - Long Range, AWD",
				ModelYear:          "2022",
				ManufacturingPlant: "Shanghai, China",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "Cybertruck from Austin",
			vin:     "7SACAEED1RA123456", // A=body LHD, E=fuel electric, D=Dual Motor AWD
			wantNil: false,
			expected: &VINInfo{
				VIN:                "7SACAEED1RA123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Austin, TX, USA",
				Model:              "Cybertruck",
				BodyType:           "Pickup, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - AWD",
				ModelYear:          "2024",
				ManufacturingPlant: "Austin, TX, USA",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "Unknown manufacturer",
			vin:     "WVWZZZ1KZAW123456",
			wantNil: false,
			expected: &VINInfo{
				VIN:                "WVWZZZ1KZAW123456",
				Manufacturer:       "Unknown",
				ManufactureRegion:  "Unknown",
				Model:              "Unknown",
				BodyType:           "Unknown",
				FuelType:           "Unknown",
				Powertrain:         "Unknown",
				ModelYear:          "Unknown",
				ManufacturingPlant: "Unknown",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "VIN with lowercase - should normalize",
			vin:     "5yj3aaee1lf123456",
			wantNil: false,
			expected: &VINInfo{
				VIN:                "5YJ3AAEE1LF123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Fremont, CA / Austin, TX, USA",
				Model:              "Model 3",
				BodyType:           "Sedan 4-door, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - Long Range, AWD",
				ModelYear:          "2020",
				ManufacturingPlant: "Fremont, CA, USA",
				SerialNumber:       "123456",
			},
		},
		{
			name:    "VIN with whitespace - should trim",
			vin:     "  5YJ3AAEE1LF123456  ",
			wantNil: false,
			expected: &VINInfo{
				VIN:                "5YJ3AAEE1LF123456",
				Manufacturer:       "Tesla, Inc.",
				ManufactureRegion:  "Fremont, CA / Austin, TX, USA",
				Model:              "Model 3",
				BodyType:           "Sedan 4-door, LHD",
				FuelType:           "Electric",
				Powertrain:         "Dual Motor - Long Range, AWD",
				ModelYear:          "2020",
				ManufacturingPlant: "Fremont, CA, USA",
				SerialNumber:       "123456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeVIN(tt.vin)

			if tt.wantNil {
				if got != nil {
					t.Errorf("DecodeVIN() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("DecodeVIN() = nil, want non-nil")
			}

			if got.VIN != tt.expected.VIN {
				t.Errorf("VIN = %v, want %v", got.VIN, tt.expected.VIN)
			}
			if got.Manufacturer != tt.expected.Manufacturer {
				t.Errorf("Manufacturer = %v, want %v", got.Manufacturer, tt.expected.Manufacturer)
			}
			if got.ManufactureRegion != tt.expected.ManufactureRegion {
				t.Errorf("ManufactureRegion = %v, want %v", got.ManufactureRegion, tt.expected.ManufactureRegion)
			}
			if got.Model != tt.expected.Model {
				t.Errorf("Model = %v, want %v", got.Model, tt.expected.Model)
			}
			if got.BodyType != tt.expected.BodyType {
				t.Errorf("BodyType = %v, want %v", got.BodyType, tt.expected.BodyType)
			}
			if got.FuelType != tt.expected.FuelType {
				t.Errorf("FuelType = %v, want %v", got.FuelType, tt.expected.FuelType)
			}
			if got.Powertrain != tt.expected.Powertrain {
				t.Errorf("Powertrain = %v, want %v", got.Powertrain, tt.expected.Powertrain)
			}
			if got.ModelYear != tt.expected.ModelYear {
				t.Errorf("ModelYear = %v, want %v", got.ModelYear, tt.expected.ModelYear)
			}
			if got.ManufacturingPlant != tt.expected.ManufacturingPlant {
				t.Errorf("ManufacturingPlant = %v, want %v", got.ManufacturingPlant, tt.expected.ManufacturingPlant)
			}
			if got.SerialNumber != tt.expected.SerialNumber {
				t.Errorf("SerialNumber = %v, want %v", got.SerialNumber, tt.expected.SerialNumber)
			}
		})
	}
}

func TestDecodeVIN_AllModelYears(t *testing.T) {
	yearTests := []struct {
		char byte
		year string
	}{
		{'E', "2014"},
		{'F', "2015"},
		{'G', "2016"},
		{'H', "2017"},
		{'J', "2018"},
		{'K', "2019"},
		{'L', "2020"},
		{'M', "2021"},
		{'N', "2022"},
		{'P', "2023"},
		{'R', "2024"},
		{'S', "2025"},
		{'T', "2026"},
	}

	for _, tt := range yearTests {
		t.Run("Year "+tt.year, func(t *testing.T) {
			// Create a VIN with the specific year character (position 10)
			vin := "5YJ3E1EA1" + string(tt.char) + "F123456"
			got := DecodeVIN(vin)
			if got == nil {
				t.Fatal("DecodeVIN() = nil")
			}
			if got.ModelYear != tt.year {
				t.Errorf("ModelYear = %v, want %v", got.ModelYear, tt.year)
			}
		})
	}
}

func TestDecodeVIN_AllManufacturingPlants(t *testing.T) {
	plantTests := []struct {
		char  byte
		plant string
	}{
		{'F', "Fremont, CA, USA"},
		{'A', "Austin, TX, USA"},
		{'C', "Shanghai, China"},
		{'B', "Berlin, Germany"},
		{'P', "Palo Alto, CA, USA"},
		{'N', "Reno, NV, USA"},
	}

	for _, tt := range plantTests {
		t.Run("Plant "+tt.plant, func(t *testing.T) {
			// Create a VIN with the specific plant character (position 11)
			vin := "5YJ3E1EA1L" + string(tt.char) + "123456"
			got := DecodeVIN(vin)
			if got == nil {
				t.Fatal("DecodeVIN() = nil")
			}
			if got.ManufacturingPlant != tt.plant {
				t.Errorf("ManufacturingPlant = %v, want %v", got.ManufacturingPlant, tt.plant)
			}
		})
	}
}

func TestDecodeVIN_AllModels(t *testing.T) {
	modelTests := []struct {
		char  byte
		model string
	}{
		{'S', "Model S"},
		{'3', "Model 3"},
		{'X', "Model X"},
		{'Y', "Model Y"},
		{'C', "Cybertruck"},
		{'R', "Roadster"},
		{'T', "Semi"},
	}

	for _, tt := range modelTests {
		t.Run("Model "+tt.model, func(t *testing.T) {
			// Create a VIN with the specific model character (position 4)
			vin := "5YJ" + string(tt.char) + "E1EA1LF123456"
			got := DecodeVIN(vin)
			if got == nil {
				t.Fatal("DecodeVIN() = nil")
			}
			if got.Model != tt.model {
				t.Errorf("Model = %v, want %v", got.Model, tt.model)
			}
		})
	}
}
