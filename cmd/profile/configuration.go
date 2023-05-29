package profile

var AddProfileFlagKey = struct {
	ProfileName string
	AccountId   string
	Region      string
	Role        string
}{
	ProfileName: "profile-name",
	AccountId:   "account-id",
	Region:      "region",
	Role:        "role",
}

var RemoveProfileFlagKey = struct {
	ProfileName string
}{
	ProfileName: "profile-name",
}
