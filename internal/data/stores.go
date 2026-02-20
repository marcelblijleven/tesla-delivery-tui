package data

// GetStoreName returns the delivery center display name.
//
// Returns N/A when the id is 0
func GetStoreName(id string) string {
	if id == "0" {
		return "N/A"
	}
	return id
}
