import { IAdminPublicKey } from "../../interfaces/IPublicKey";
import http from "../helpers/http";

export const postPublicKey = async (data: IAdminPublicKey) => http().post("/api/sshkeys/public_keys", {
  name: data.name,
  data: data.data,
});

// eslint-disable-next-line vue/max-len
export const fetchPublicKeys = async (perPage: number, page: number) => http().get(`/api/sshkeys/public_keys?per_page=${perPage}&page=${page}`);

export const getPublicKey = async (fingerprint: string) => http().get(`/api/sshkeys/public_keys/${fingerprint}`);

export const putPublicKey = async (data: IAdminPublicKey) => http().put(`/api/sshkeys/public_keys/${data.fingerprint}`, {
  name: data.name,
  data: data.data,
});

export const removePublicKey = async (fingerprint: string) => http().delete(`/api/sshkeys/public_keys/${fingerprint}`);
