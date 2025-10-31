import { sshApi } from "@/api/http";
import { PublicKeyRequest, UpdatePublicKeyRequest } from "@/api/client/api";
import { IPublicKey, IPublicKeyCreate } from "@/interfaces/IPublicKey";

export const createPublicKey = async (data: IPublicKeyCreate) => sshApi.createPublicKey(data as PublicKeyRequest);

export const fetchPublicKeys = async (
  page: number,
  perPage: number,
  filter?: string,
) => sshApi.getPublicKeys(filter, page, perPage);

export const updatePublicKey = async (data: IPublicKey) => sshApi.updatePublicKey(data.fingerprint, data as UpdatePublicKeyRequest);

export const deletePublicKey = async (fingerprint: string) => sshApi.deletePublicKey(fingerprint);
