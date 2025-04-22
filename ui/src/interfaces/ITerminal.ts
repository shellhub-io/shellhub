export interface IConnectToTerminal {
  username: string;
  password?: string;
  privateKey?: string;
  signature?: string;
  fingerprint?: string;
}

export enum TerminalAuthMethods {
  Password = "Password",
  PrivateKey = "Private Key",
}
