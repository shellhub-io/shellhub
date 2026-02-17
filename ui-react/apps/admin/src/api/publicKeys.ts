import apiClient from "./client";
import { PublicKey, PublicKeyFilter } from "../types/publicKey";
import { PaginatedResponse } from "../types/api";

export async function getPublicKeys(
  page = 1,
  perPage = 10,
): Promise<PaginatedResponse<PublicKey>> {
  const res = await apiClient.get<PublicKey[]>("/api/sshkeys/public-keys", {
    params: { page, per_page: perPage },
  });
  const totalCount = parseInt(res.headers["x-total-count"] || "0", 10);
  return { data: res.data, totalCount };
}

export async function createPublicKey(payload: {
  name: string;
  data: string;
  username: string;
  filter: PublicKeyFilter;
}): Promise<PublicKey> {
  const { data } = await apiClient.post<PublicKey>(
    "/api/sshkeys/public-keys",
    payload,
  );
  return data;
}

export async function updatePublicKey(
  fingerprint: string,
  payload: { name: string; username: string; filter: PublicKeyFilter },
): Promise<void> {
  await apiClient.put(`/api/sshkeys/public-keys/${fingerprint}`, payload);
}

export async function deletePublicKey(fingerprint: string): Promise<void> {
  await apiClient.delete(`/api/sshkeys/public-keys/${fingerprint}`);
}
