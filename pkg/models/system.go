package models

type System struct {
	Setup bool `json:"setup"`
	// Authentication manages the settings for available authentication methods, such as manual
	// username/password authentication and SAML authentication. Each authentication method
	// can be individually enabled or disabled.
	Authentication *SystemAuthentication `json:"authentication" bson:"authentication"`
}

type SystemAuthentication struct {
	Local *SystemAuthenticationLocal `json:"local" bson:"local"`
	SAML  *SystemAuthenticationSAML  `json:"saml" bson:"saml"`
}

type SystemAuthenticationLocal struct {
	// Enabled indicates whether manual authentication using a username and password is enabled or
	// not.
	Enabled bool `json:"enabled" bool:"enabled"`
}

type SystemAuthenticationSAML struct {
	// Enabled indicates whether SAML authentication is enabled.
	Enabled bool           `json:"enabled" bson:"enabled"`
	Idp     *SystemIdpSAML `json:"idp" bson:"idp"`
	Sp      *SystemSpSAML  `json:"sp" bson:"sp"`
}

type SystemAuthenticationBinding struct {
	Post     string `json:"post" bson:"post"`
	Redirect string `json:"redirect" bson:"redirect"`
	// PreferredBinding defines the preferred SAML binding method.
	Preferred string `json:"preferred" bson:"preferred"`
}

type SystemIdpSAML struct {
	EntityID string                       `json:"entity_id" bson:"entity_id"`
	Binding  *SystemAuthenticationBinding `json:"binding" bson:"binding"`
	// Certificates is a list of X.509 certificates provided by the IdP. These certificates are used
	// by the SP to validate the digital signatures of SAML assertions issued by the IdP.
	Certificates []string `json:"certificates" bson:"certificates"`
	// Mappings defines how IdP SAML attributes map to ShellHub attributes.
	//
	// Example:
	// 	{
	//		"external_id": "user_id",
	// 		"email": "emailaddress",
	// 		"name": "displayName"
	// 	}
	Mappings map[string]string `json:"mappings" bson:"mappings"`
}

type SystemSpSAML struct {
	// SignRequests indicates whether ShellHub should sign authentication requests.
	// If enabled, an X509 certificate is used to sign the request, and the IdP must authenticate
	// the request using the corresponding public certificate. Enabling this option disables
	// the "IdP-initiated" authentication pipeline.
	SignAuthRequests bool `json:"sign_auth_requests" bson:"sign_auth_requests"`
	// Certificate is an X509 certificate that the IdP uses to verify the authenticity of the
	// authentication request signed by ShellHub. This certificate corresponds to the private key
	// in the [SystemSpSAML.PrivateKey] and it is only populated when [SystemSpSAML.SignAuthRequests]
	// is true.
	Certificate string `json:"certificate" bson:"certificate"`
	// PrivateKey is an encrypted private key used by ShellHub to sign authentication requests.
	// The IdP verifies the signature using the [SystemSpSAML.Certificate]. It is only populated
	// when [SystemSpSAML.SignAuthRequests] is true.
	PrivateKey string `json:"-" bson:"private_key"`
}

type SAMLBinding struct {
	URL     string
	Binding string
}

const (
	SAMLBindingPost     = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
	SAMLBindingRedirect = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
)

func (s *SystemIdpSAML) GetBinding() SAMLBinding {
	if s.Binding.Preferred == "post" {
		return SAMLBinding{URL: s.Binding.Post, Binding: SAMLBindingPost}
	} else if s.Binding.Preferred == "redirect" {
		return SAMLBinding{URL: s.Binding.Redirect, Binding: SAMLBindingRedirect}
	} else if s.Binding.Post != "" {
		return SAMLBinding{URL: s.Binding.Post, Binding: SAMLBindingPost}
	} else {
		return SAMLBinding{URL: s.Binding.Redirect, Binding: SAMLBindingRedirect}
	}
}
