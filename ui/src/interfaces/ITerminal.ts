export interface IConnectToTerminal {
  username: string;
  password?: string;
  privateKey?: string;
  passphrase?: string;
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
  passphrase?: string;
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
  data: string; // Input data, limited to 4096 characters
}

export interface SignatureMessage {
  kind: MessageKind.Signature;
  data: string; // base64-encoded challenge or response
}

export interface ErrorMessage {
  kind: MessageKind.Error;
  data: string; // Error Message Content
}

export interface ITerminalTheme {
  name: string;
  description: string;
  colors: {
    background: string;
    foreground: string;
    cursor?: string;
    selection?: string;
    black?: string;
    red?: string;
    green?: string;
    yellow?: string;
    blue?: string;
    magenta?: string;
    cyan?: string;
    white?: string;
    brightBlack?: string;
    brightRed?: string;
    brightGreen?: string;
    brightYellow?: string;
    brightBlue?: string;
    brightMagenta?: string;
    brightCyan?: string;
    brightWhite?: string;
  };
}

export interface IThemeMetadata {
  name: string;
  file: string;
  dark: boolean;
  preview: {
    background: string;
    foreground: string;
  };
}
