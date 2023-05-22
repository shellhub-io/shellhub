import http from "../helpers/http";
import { sshApi } from "../../api/http";
import { IPublicKey } from "@/interfaces/IPublicKey";

export const postPublicKey = async (data : any | IPublicKey) => sshApi.createPublicKey(data);

export const fetchPublicKeys = async (
  page : number,
  perPage: number,
  filter : string,
) => {
  if (filter) return sshApi.getPublicKeys(filter, page, perPage);

  return sshApi.getPublicKeys(filter, page, perPage);
};

export const getPublicKey = async (fingerprint : string) => http().get(`/sshkeys/public-keys/${fingerprint}`); // TODO

export const putPublicKey = async (data : any | IPublicKey) => sshApi.updatePublicKey(data.fingerprint, data);

export const removePublicKey = async (fingerprint : string) => sshApi.deletePublicKey(fingerprint);
