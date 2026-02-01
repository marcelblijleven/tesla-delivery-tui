package model

import "strings"

// VINInfo contains decoded VIN information
type VINInfo struct {
	VIN               string
	Manufacturer      string
	ManufactureRegion string
	Model             string
	BodyType          string
	FuelType          string
	Powertrain        string
	ModelYear         string
	ManufacturingPlant string
	SerialNumber      string
}

// World Manufacturer Identifier (first 3 characters)
var wmiMap = map[string]struct {
	Manufacturer string
	Region       string
}{
	"5YJ": {"Tesla, Inc.", "Fremont, CA / Austin, TX, USA"},
	"7SA": {"Tesla, Inc.", "Austin, TX, USA"},
	"7G2": {"Tesla, Inc.", "Reno, NV, USA"},
	"LRW": {"Tesla, Inc.", "Shanghai, China"},
	"XP7": {"Tesla, Inc.", "Berlin, Germany"},
}

// Model codes (4th character)
var modelMap = map[byte]string{
	'S': "Model S",
	'3': "Model 3",
	'X': "Model X",
	'Y': "Model Y",
	'C': "Cybertruck",
	'R': "Roadster",
	'T': "Semi",
}

// Body type (5th character) - varies by model
var bodyTypeMap = map[string]map[byte]string{
	"S": {
		'A': "Hatchback 5-door, LHD",
		'B': "Hatchback 5-door, RHD",
	},
	"3": {
		'A': "Sedan 4-door, LHD",
		'B': "Sedan 4-door, RHD",
	},
	"X": {
		'A': "SUV 5-door, LHD",
		'B': "SUV 5-door, RHD",
	},
	"Y": {
		'A': "SUV 5-door, LHD",
		'B': "SUV 5-door, RHD",
		'C': "SUV 5-door, LHD",
		'D': "SUV 5-door, RHD",
		'E': "SUV 5-door, LHD",
		'F': "SUV 5-door, RHD",
	},
	"C": {
		'A': "Pickup, LHD",
		'B': "Pickup, RHD",
	},
}

// Fuel type (7th character)
var fuelTypeMap = map[byte]string{
	'E': "Electric",
}

// Powertrain (8th character) - varies by model
var powertrainMapS = map[byte]string{
	'1': "Single Motor - Standard",
	'2': "Dual Motor - Standard",
	'3': "Dual Motor - Performance",
	'4': "Dual Motor - Standard",
	'5': "Dual Motor - Performance",
	'6': "Tri Motor",
	'C': "Base, Standard Range",
	'D': "Base, Long Range",
}

var powertrainMap3 = map[byte]string{
	'A': "Single Motor - Standard Range Plus, RWD",
	'B': "Single Motor - Standard Range, RWD",
	'C': "Single Motor - Standard Range Plus, RWD",
	'D': "Single Motor - Mid Range, RWD",
	'E': "Dual Motor - Long Range, AWD",
	'F': "Dual Motor - Performance, AWD",
	'G': "Single Motor - Standard Range Plus, RWD",
	'H': "Single Motor - Standard Range Plus, RWD",
	'K': "Single Motor - Standard Range Plus, RWD",
	'L': "Dual Motor - Long Range, AWD",
	'N': "Dual Motor - Long Range, AWD",
	'P': "Dual Motor - Performance, AWD",
	'Q': "Dual Motor - Long Range, AWD",
	'R': "Dual Motor - Performance, AWD",
}

var powertrainMapY = map[byte]string{
	'A': "Single Motor - Standard Range, RWD",
	'C': "Dual Motor - Long Range, AWD",
	'D': "Dual Motor - Long Range, AWD",
	'E': "Dual Motor - Performance, AWD",
	'F': "Dual Motor - Long Range, AWD",
	'G': "Dual Motor - Performance, AWD",
	'H': "Dual Motor - Long Range, AWD",
	'J': "Single Motor - RWD",
	'W': "Single Motor - RWD",
}

var powertrainMapC = map[byte]string{
	'D': "Dual Motor - AWD",
	'E': "Tri Motor - AWD",
}

// Model year (10th character)
var yearMap = map[byte]string{
	'E': "2014",
	'F': "2015",
	'G': "2016",
	'H': "2017",
	'J': "2018",
	'K': "2019",
	'L': "2020",
	'M': "2021",
	'N': "2022",
	'P': "2023",
	'R': "2024",
	'S': "2025",
	'T': "2026",
}

// Manufacturing plant (11th character)
var plantMap = map[byte]string{
	'F': "Fremont, CA, USA",
	'A': "Austin, TX, USA",
	'C': "Shanghai, China",
	'B': "Berlin, Germany",
	'P': "Palo Alto, CA, USA",
	'N': "Reno, NV, USA",
}

// DecodeVIN decodes a Tesla VIN into its component parts
func DecodeVIN(vin string) *VINInfo {
	vin = strings.ToUpper(strings.TrimSpace(vin))

	if len(vin) != 17 {
		return nil
	}

	info := &VINInfo{
		VIN: vin,
	}

	// WMI (chars 1-3)
	wmi := vin[0:3]
	if data, ok := wmiMap[wmi]; ok {
		info.Manufacturer = data.Manufacturer
		info.ManufactureRegion = data.Region
	} else {
		info.Manufacturer = "Unknown"
		info.ManufactureRegion = "Unknown"
	}

	// Model (char 4)
	if model, ok := modelMap[vin[3]]; ok {
		info.Model = model
	} else {
		info.Model = "Unknown"
	}

	// Body type (char 5) - depends on model
	modelKey := string(vin[3])
	if bodyTypes, ok := bodyTypeMap[modelKey]; ok {
		if body, ok := bodyTypes[vin[4]]; ok {
			info.BodyType = body
		} else {
			info.BodyType = "Unknown"
		}
	} else {
		info.BodyType = "Unknown"
	}

	// Fuel type (char 7)
	if fuel, ok := fuelTypeMap[vin[6]]; ok {
		info.FuelType = fuel
	} else {
		info.FuelType = "Unknown"
	}

	// Powertrain (char 8) - depends on model
	switch vin[3] {
	case 'S':
		if pt, ok := powertrainMapS[vin[7]]; ok {
			info.Powertrain = pt
		} else {
			info.Powertrain = "Unknown"
		}
	case '3':
		if pt, ok := powertrainMap3[vin[7]]; ok {
			info.Powertrain = pt
		} else {
			info.Powertrain = "Unknown"
		}
	case 'Y':
		if pt, ok := powertrainMapY[vin[7]]; ok {
			info.Powertrain = pt
		} else {
			info.Powertrain = "Unknown"
		}
	case 'C':
		if pt, ok := powertrainMapC[vin[7]]; ok {
			info.Powertrain = pt
		} else {
			info.Powertrain = "Unknown"
		}
	default:
		info.Powertrain = "Unknown"
	}

	// Model year (char 10)
	if year, ok := yearMap[vin[9]]; ok {
		info.ModelYear = year
	} else {
		info.ModelYear = "Unknown"
	}

	// Manufacturing plant (char 11)
	if plant, ok := plantMap[vin[10]]; ok {
		info.ManufacturingPlant = plant
	} else {
		info.ManufacturingPlant = "Unknown"
	}

	// Serial number (chars 12-17)
	info.SerialNumber = vin[11:17]

	return info
}
