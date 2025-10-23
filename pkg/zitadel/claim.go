package zitadel_pkg

type ZitadelClaim struct {
	AMR                          string                       `json:"amr"`
	Aud                          string                       `json:"aud"`
	Exp                          int64                        `json:"exp"`
	Iat                          int64                        `json:"iat"`
	Iss                          string                       `json:"iss"`
	Sub                          string                       `json:"sub"`
	Azp                          string                       `json:"azp"`
	Email                        string                       `json:"email"`
	EmailVerified                string                       `json:"email_verified"`
	FamilyName                   string                       `json:"family_name"`
	GivenName                    string                       `json:"given_name"`
	PreferredUsername            string                       `json:"preferred_username"`
	UrnZitadelIAMOrgProjectRoles map[string]map[string]string `json:"urn:zitadel:iam:org:project:roles"`
	Metadata                     map[string]string            `json:"urn:zitadel:iam:user:metadata"`
	IdToken                      string                       `json:"id_token"`
	Token                        string                       `json:"token"`
}
