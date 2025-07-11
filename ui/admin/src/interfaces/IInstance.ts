export interface IAdminSAMLConfig {
  enable: boolean,
  idp: {
    metadata_url?: string,
    entity_id?: string,
    signon_url?: string,
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
