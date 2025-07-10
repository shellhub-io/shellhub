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

export interface LoginFormData {
  username: string;
  password: string;
  authenticationMethod: TerminalAuthMethods;
  privateKey?: string;
}

export interface WebTermDimensions {
  cols: number;
  rows: number;
}

export enum MessageKind {
  Input = 1,
  Resize,
  Signature,
  Error,
}

export interface ResizeMessage {
  kind: MessageKind.Resize;
  data: WebTermDimensions;
}

export interface InputMessage {
  kind: MessageKind.Input;
  data: number[];
}

export interface SignatureMessage {
  kind: MessageKind.Signature;
  data: string; // base64-encoded challenge or response
}

export interface ErrorMessage {
  kind: MessageKind.Error;
  data: string; // Error Message Content
}
