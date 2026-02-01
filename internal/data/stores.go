package data

// LookupStoreName returns the delivery center display name.
//
// The Tesla API scheduling task typically returns a human-readable name
// in the deliveryAddressTitle field. This function passes it through
// unchanged, with a special case for "0" which the API uses for unassigned.
func LookupStoreName(id string) string {
	if id == "0" {
		return "N/A"
	}
	return id
}
