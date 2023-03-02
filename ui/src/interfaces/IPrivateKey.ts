export interface IPrivateKey {
  name: string;
  data: string;
}

export interface IPrivateKeyError extends Error {
  code: number;
  body: string[];
}
