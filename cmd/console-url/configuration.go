package consoleurl

import v1 "github.com/hunoz/maroon-api/api/v1"

var FlagKey = struct {
	AccountId  string
	AccessType v1.AccessType
	Duration   string
	Token      string
}{
	AccountId:  "account-id",
	AccessType: "access-type",
	Duration:   "duration",
	Token:      "token",
}
