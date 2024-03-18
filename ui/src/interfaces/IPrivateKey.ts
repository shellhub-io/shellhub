export interface IPrivateKey {
  id: number,
  name: string;
  data: string;
}

export interface IPrivateKeyError extends Error {
  code: number;
  body: string[];
}
