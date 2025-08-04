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

export interface IAdminSAMLConfig {
  enable: boolean,
  idp: {
    metadata_url?: string,
    entity_id?: string,
    binding?: SAMLBinding,
    certificate?: string
    mappings?: {
      email: string,
      name: string,
    }
  },
  sp: {
    sign_requests: boolean
  }
}
