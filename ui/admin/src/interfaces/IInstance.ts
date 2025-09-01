// Binding needs to receive either post or redirect URL, or both.
type SAMLBinding = ({
  post: string,
} | {
  redirect: string,
} | {
  post: string,
  redirect: string,
}) & {
  preferred?: "post" | "redirect",
}

export interface IAdminUpdateSAML {
  enable: boolean;
  idp: {
    metadata_url?: string,
    entity_id?: string,
    binding?: SAMLBinding,
    certificate?: string,
    mappings?: {
      email: string,
      name: string,
    },
  };
  sp: {
    sign_requests?: boolean
  }
}

export interface IAdminSAML {
  enabled: boolean,
  idp: {
    entity_id: string,
    binding: SAMLBinding,
    certificates: Array<string>,
    mappings?: {
      email: string,
      name: string,
    }
  },
  sp: {
    certificate: string
    sign_auth_requests: boolean
  },
  auth_url: string,
  assertion_url: string,
}

export interface IAdminAuth {
  local: { enabled: boolean };
  saml: IAdminSAML;
}
