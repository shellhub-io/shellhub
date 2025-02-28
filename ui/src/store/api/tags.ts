import { UpdateTagRequest } from "@/api/client";
import { tagsApi } from "../../api/http";

export const createTag = async (tenant: string, name: string) => tagsApi.createTag(tenant, { name });

export const updateTag = async (
  tenant: string,
  currentName: string,
  newName: UpdateTagRequest,
) => tagsApi.updateTag(tenant, currentName, newName);

export const removeTag = async (tenant: string, currentName: string) => tagsApi.deleteTag(tenant, currentName);

export const removeTagFromDevice = async (tenant: string, uid: string, name: string) => tagsApi.pullTagFromDevice(tenant, uid, name);

export const removeTagFromFirewallRule = async (tenant: string, name: string) => tagsApi.pullTagFromFirewallRule(tenant, name);

export const removeTagFromPublicKey = async (
  tenant: string,
  fingerprint: string,
  name: string,
) => tagsApi.pullTagFromPublicKey(tenant, fingerprint, name);

export const pushTagToDevice = async (tenant: string, uid: string, name: string) => tagsApi.pushTagToDevice(tenant, uid, name);

export const pushTagToFirewallRule = async (tenant: string, name: string) => tagsApi.pushTagToFirewallRule(tenant, name);

export const pushTagToPublicKey = async (
  tenant: string,
  fingerprint: string,
  name: string,
) => tagsApi.pushTagToPublicKey(tenant, fingerprint, name);

export const getTags = async (
  tenant: string,
  filter : string | undefined,
  page : number,
  perPage: number,
) => tagsApi.getTags(tenant, filter, page, perPage);
