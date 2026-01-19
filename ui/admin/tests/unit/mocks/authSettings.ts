export const mockAuthSettings = {
  local: {
    enabled: true,
  },
  saml: {
    enabled: true,
    auth_url: "https://auth.example.com/sso",
    assertion_url: "http://localhost:3000/api/user/saml/auth",
    idp: {
      entity_id: "https://idp.example.com/entity",
      binding: {
        post: "https://idp.example.com/sso/post",
        redirect: "https://idp.example.com/sso/redirect",
      },
      certificates: ["-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKZQ..."],
      mappings: {
        email: "emailAddress",
        name: "displayName",
      },
    },
    sp: {
      sign_auth_requests: true,
      certificate: "-----BEGIN CERTIFICATE-----\nSP_CERT_CONTENT\n-----END CERTIFICATE-----",
    },
  },
};

export const mockAuthSettingsLocalOnly = {
  local: {
    enabled: true,
  },
  saml: {
    enabled: false,
    auth_url: "",
    assertion_url: "",
    idp: {
      entity_id: "",
      binding: {
        post: "",
        redirect: "",
      },
      certificates: [],
      mappings: {
        email: "",
        name: "",
      },
    },
    sp: {
      sign_auth_requests: false,
      certificate: "",
    },
  },
};
