import { Filter } from "@/interfaces/IFilter";

export interface IPublicKey {
  created_at: string;
  data: string;
  filter: Filter;
  fingerprint: string;
  name: string;
  tenant_id: string;
  username: string;
}

export type IPublicKeyCreate = Pick<IPublicKey, "data" | "name" | "username" | "filter">
