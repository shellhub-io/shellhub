export interface IConnectToTerminal {
  password?: string;
  signature?: string;
  fingerprint?: string;
}

export interface ITerminalFrames {
  incMessage: string;
  incTime: number;
}

export interface ITerminalLog {
  uid: string;
  message: string;
  tenant_id: string;
  time: string;
  width: number;
  height: number;
}
