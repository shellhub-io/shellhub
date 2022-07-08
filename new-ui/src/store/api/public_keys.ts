import http from '../helpers/http';
import { sshApi } from "../../api/http";

export const postPublicKey = async (data : any) => sshApi.createPublicKey(data);

export const fetchPublicKeys = async (
    page : any,
    perPage: any,
    filter : any,
  ) => {
    if (filter) return sshApi.getPublicKeys(filter, page, perPage);
  
    return sshApi.getPublicKeys(filter, page, perPage);
  };

export const getPublicKey = async (fingerprint : any) => http().get(`/sshkeys/public-keys/${fingerprint}`); // TODO

export const putPublicKey = async (data : any) => sshApi.updatePublicKey(data.fingerprint, data);

export const removePublicKey = async (fingerprint : any) => sshApi.deletePublicKey(fingerprint);
