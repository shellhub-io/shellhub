import { IPublicKey } from "../../interfaces/IPublicKey";
import http from "../helpers/http";

export const postPublicKey = async (data: IPublicKey) => http().post("/api/sshkeys/public_keys", {
  name: data.name,
  data: data.data,
});

// eslint-disable-next-line vue/max-len
export const fetchPublicKeys = async (perPage: number, page: string) => http().get(`/api/sshkeys/public_keys?per_page=${perPage}&page=${page}`);

export const getPublicKey = async (fingerprint: string) => http().get(`/api/sshkeys/public_keys/${fingerprint}`);

export const putPublicKey = async (data: IPublicKey) => http().put(`/api/sshkeys/public_keys/${data.fingerprint}`, {
  name: data.name,
  data: data.data,
});

export const removePublicKey = async (fingerprint: string) => http().delete(`/api/sshkeys/public_keys/${fingerprint}`);
