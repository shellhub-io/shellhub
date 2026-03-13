export interface SystemInfo {
  version: string;
  setup: boolean;
  authentication: {
    local: boolean;
    saml: boolean;
  };
  endpoints: {
    ssh: string;
    api: string;
  };
}

export interface SetupRequest {
  name: string;
  username: string;
  email: string;
  password: string;
}
