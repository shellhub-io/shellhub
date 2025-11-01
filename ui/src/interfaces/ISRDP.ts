/**
 * SRDP (Secure Screen Sharing Protocol) Client Interfaces
 * 
 * These interfaces define the data structures for the SRDP/RFB client
 * that connects to remote devices using WebAssembly.
 */

export interface IConnectToSRDP {
  device: string;
  username: string;
  password: string;
}

export interface SRDPLoginFormData {
  device: string;
  username: string;
  password?: string;
  privateKey?: string;
  passphrase?: string;
  display?: string;
}

export enum SRDPAuthMethod {
  Password = "Password",
  PrivateKey = "Private Key",
}

export interface SRDPConnectionParams {
  device: string;
  username: string;
  password: string;
  authenticationMethod: SRDPAuthMethod;
}

/**
 * SRDP WebSocket Message Types
 * These match the message kinds used in the SRDP protocol
 */
export enum SRDPMessageKind {
  Success = 5, // Successful connection message
  Error = 4,   // Error message from server
}

export interface SRDPMessage {
  kind: SRDPMessageKind;
  data?: unknown;
}

export interface SRDPErrorMessage {
  kind: SRDPMessageKind.Error;
  data: string;
}

export interface SRDPSuccessMessage {
  kind: SRDPMessageKind.Success;
  data?: unknown;
}

/**
 * SRDP Client State
 */
export enum SRDPClientState {
  Disconnected = "disconnected",
  Connecting = "connecting",
  Authenticating = "authenticating",
  Connected = "connected",
  Error = "error",
}

/**
 * SRDP Error Types
 */
export interface SRDPError {
  message: string;
  code?: string;
  details?: unknown;
}
