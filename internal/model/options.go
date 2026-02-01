package model

import "strings"

// TeslaOptionCodes maps option codes to human-readable descriptions
var TeslaOptionCodes = map[string]string{
	// Autopilot & FSD
	"APBS": "Autopilot - Basic",
	"APF0": "Autopilot - Basic (different iteration)",
	"APF1": "Autopilot - Enhanced",
	"APF2": "Full Self-Driving Capability",
	"APH0": "Autopilot 2.0 Hardware",
	"APH1": "Autopilot 2.0 Hardware",
	"APH2": "Autopilot 2.5 Hardware",
	"APH3": "Autopilot 3.0 Hardware (HW3)",
	"APH4": "Autopilot 4.0 Hardware (HW4)",
	"APPA": "Autopilot Active Safety Features",
	"APPB": "Enhanced Autopilot",
	"APPF": "Full Self-Driving Capability",

	// Paint Colors
	"PBSB": "Solid Black",
	"PBCW": "Solid Black",
	"PMSS": "Silver Metallic",
	"PMTG": "Midnight Silver Metallic",
	"PPMR": "Red Multi-Coat",
	"PPSB": "Obsidian Black Metallic",
	"PPSR": "Signature Red",
	"PPSW": "Pearl White Multi-Coat",
	"PPTI": "Titanium Metallic",
	"PMNG": "Midnight Cherry Red",
	"PMBL": "Ultra Blue",
	"PN00": "Midnight Silver Metallic",
	"PN01": "Solid Black",
	"PR00": "Pearl White Multi-Coat",
	"PR01": "Solid Black",
	"PMAB": "Quicksilver",
	"PMSG": "Stealth Grey",
	"PMMB": "Ultra Blue",

	// Interior
	"IBB0": "All Black Interior",
	"IBB1": "All Black Interior",
	"IBE0": "Black & White Interior",
	"IBW0": "Black & White Interior",
	"ICW0": "Cream Interior",
	"IPB0": "Black Premium Interior",
	"IPB1": "Black Premium Interior",
	"IPB11": "Black Premium Interior",
	"IPW0": "White Premium Interior",
	"IPW1": "White Premium Interior",
	"IWW0": "White Interior",
	"IBC0": "Black Interior",
	"IN3BB": "All Black Premium Interior",
	"IN3BW": "Black and White Premium Interior",
	"IN3PB": "Black Premium Interior",
	"IN3PW": "White Premium Interior",

	// Battery & Range
	"BT37": "75 kWh Battery",
	"BT40": "40 kWh Battery",
	"BT60": "60 kWh Battery",
	"BT70": "70 kWh Battery",
	"BT85": "85 kWh Battery",
	"BTX4": "90 kWh Battery",
	"BTX5": "75 kWh Battery",
	"BTX6": "100 kWh Battery",
	"BTX7": "75 kWh Battery",
	"BTX8": "100 kWh Battery",

	// Drive & Performance
	"DV2W": "Rear-Wheel Drive",
	"DV4W": "All-Wheel Drive (Dual Motor)",
	"DR01": "Rear-Wheel Drive",
	"DR02": "All-Wheel Drive (Dual Motor)",
	"DRRH": "Rear-Wheel Drive",
	"DRRL": "Rear-Wheel Drive Long Range",
	"MDL3": "Model 3",
	"MDLS": "Model S",
	"MDLX": "Model X",
	"MDLY": "Model Y",
	"REEU": "European Region",
	"RENA": "North American Region",
	"RENC": "Canadian Region",

	// Wheels
	"W32P": "20\" Performance Wheels",
	"W32D": "20\" Gray Performance Wheels",
	"W33D": "20\" Gray Wheels",
	"W38B": "18\" Aero Wheels",
	"W39B": "19\" Sport Wheels",
	"W40B": "18\" Aero Wheels",
	"W41B": "19\" Gemini Wheels",
	"WR00": "Wheel Upgrade",
	"WR01": "19\" Wheels",
	"WS10": "21\" Arachnid Wheels",
	"WS90": "19\" Tempest Wheels",
	"WT19": "19\" Wheels",
	"WT20": "20\" Wheels",
	"WY18B": "18\" Aero Wheels",
	"WY19B": "19\" Gemini Wheels",
	"WY19P": "19\" Sport Wheels",
	"WY20P": "20\" Induction Wheels",
	"WY21P": "21\" Überturbine Wheels",

	// Seats
	"ST00": "Non-Performance Seats",
	"ST01": "Performance Seats",
	"ST0Y": "Standard Seats",
	"ST31": "Performance Seats with Lumbar",
	"STY5S": "5 Seat Interior",
	"STY7S": "7 Seat Interior",

	// Tow Hitch
	"TW00": "No Tow Hitch",
	"TW01": "Tow Hitch",
	"TW02": "Tow Hitch",

	// Charging
	"CH00": "Standard Charging",
	"CH01": "Dual Chargers",
	"CH04": "72 Amp Charger",
	"CH05": "48 Amp Charger",
	"CH07": "48 Amp Charger",
	"SC00": "No Supercharging",
	"SC01": "Supercharging Enabled",
	"SC04": "Pay Per Use Supercharging",
	"SC05": "Free Unlimited Supercharging",

	// Roof
	"RF3G": "Glass Roof",
	"RFFG": "Fixed Glass Roof",
	"RFPX": "Panoramic Sunroof",
	"RFP0": "All Glass Panoramic Roof",
	"RFP2": "Sunroof",

	// Cold Weather
	"CW00": "No Cold Weather Package",
	"CW02": "Cold Weather Package (Subzero)",
	"CPF0": "Standard Connectivity",
	"CPF1": "Premium Connectivity",

	// Model Y Specific
	"MTY01": "Model Y Standard Range",
	"MTY03": "Model Y Long Range",
	"MTY04": "Model Y Performance",
	"MTY05": "Model Y Long Range AWD",
	"MTY07": "Model Y Long Range RWD",
	"MTY12": "Model Y AWD",
	"MTY52": "Model Y Long Range AWD",

	// Model 3 Specific
	"MT300": "Model 3 Standard Range Plus",
	"MT301": "Model 3 Standard Range Plus",
	"MT302": "Model 3 Long Range",
	"MT303": "Model 3 Long Range AWD",
	"MT304": "Model 3 Long Range Performance",
	"MT305": "Model 3 Standard Range Plus",
	"MT307": "Model 3 Long Range AWD",
	"MT308": "Model 3 Performance",
	"MT310": "Model 3 Long Range",
	"MT314": "Model 3 Standard Range RWD",
	"MT315": "Model 3 Long Range RWD",
	"MT316": "Model 3 Long Range AWD",
	"MT317": "Model 3 Performance AWD",
	"MT336": "Model 3 Standard Range RWD",
	"MT337": "Model 3 Long Range AWD",

	// Misc
	"AD02": "NEMA 14-50 Adapter",
	"AD15": "Power Adapter",
	"GLFR": "Gloss Finish",
	"HL31": "Head Lights",
	"HL32": "Matrix LED Headlights",
	"HP00": "No Heat Pump",
	"HP01": "Heat Pump",
	"LLP1": "License Plate Bracket",
	"LLP2": "No License Plate Bracket",
	"OSSB": "Safety Belt",
	"PAF0": "No Paint Armor Film",
	"PAF1": "Paint Armor Film",
	"PI00": "No Premium Interior",
	"PI01": "Premium Interior",
	"PK00": "No Performance Package",
	"PL30": "No Rear Heated Seats",
	"PL31": "Rear Heated Seats",
	"PRM30": "Premium 30",
	"PRM31": "Premium 31",
	"PRM35": "Premium 35",
	"PS00": "No Parcel Shelf",
	"PS01": "Parcel Shelf",
	"RS3H": "Second Row Heated Seats",
	"S01B": "Black Textile Seats",
	"S02W": "White Seats",
	"SP00": "No Spoiler",
	"SP01": "Carbon Fiber Spoiler",
	"SPMR": "Red Multi-Coat",
	"SU00": "Standard Suspension",
	"SU01": "Smart Air Suspension",
	"SU03": "Performance Suspension",
	"TP01": "Tech Package",
	"TP02": "Tech Package 2",
	"TR00": "No Roof Rack",
	"TR01": "Roof Rack",
	"TRA1": "Rear-Facing Seats",
	"UM01": "Universal Mobile Connector",
	"USSB": "Safety Score Beta",
	"UTSB": "Safety Belt",
	"ZINV": "Inventory Vehicle",
}

// DecodeOptionCode returns a human-readable description for an option code
func DecodeOptionCode(code string) string {
	if desc, ok := TeslaOptionCodes[code]; ok {
		return desc
	}
	return "" // Unknown code
}

// DecodeOptions takes a comma-separated string of option codes and returns decoded options
func DecodeOptions(optionsStr string) []DecodedOption {
	if optionsStr == "" {
		return nil
	}

	codes := strings.Split(optionsStr, ",")
	var options []DecodedOption

	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		desc := DecodeOptionCode(code)
		options = append(options, DecodedOption{
			Code:        code,
			Description: desc,
		})
	}

	return options
}

// DecodedOption represents a decoded vehicle option
type DecodedOption struct {
	Code        string
	Description string
}

// CategorizeOptions groups options by category
func CategorizeOptions(options []DecodedOption) map[string][]DecodedOption {
	categories := map[string][]DecodedOption{
		"Model":       {},
		"Paint":       {},
		"Interior":    {},
		"Wheels":      {},
		"Autopilot":   {},
		"Charging":    {},
		"Other":       {},
	}

	for _, opt := range options {
		code := opt.Code
		switch {
		case strings.HasPrefix(code, "MDL") || strings.HasPrefix(code, "MT"):
			categories["Model"] = append(categories["Model"], opt)
		case strings.HasPrefix(code, "P") && (strings.HasPrefix(code, "PP") || strings.HasPrefix(code, "PM") || strings.HasPrefix(code, "PB") || strings.HasPrefix(code, "PN") || strings.HasPrefix(code, "PR")):
			categories["Paint"] = append(categories["Paint"], opt)
		case strings.HasPrefix(code, "I") || strings.HasPrefix(code, "ST"):
			categories["Interior"] = append(categories["Interior"], opt)
		case strings.HasPrefix(code, "W"):
			categories["Wheels"] = append(categories["Wheels"], opt)
		case strings.HasPrefix(code, "AP"):
			categories["Autopilot"] = append(categories["Autopilot"], opt)
		case strings.HasPrefix(code, "SC") || strings.HasPrefix(code, "CH"):
			categories["Charging"] = append(categories["Charging"], opt)
		default:
			categories["Other"] = append(categories["Other"], opt)
		}
	}

	return categories
}
