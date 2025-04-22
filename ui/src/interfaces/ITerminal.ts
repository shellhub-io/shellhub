export interface IConnectToTerminal {
  password?: string;
  signature?: string;
  fingerprint?: string;
}

export enum TerminalAuthMethods {
  Password = "Password",
  PrivateKey = "Private Key",
}
