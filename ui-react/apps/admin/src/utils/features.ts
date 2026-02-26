import { getConfig } from "../env";

export function isPremiumFeature(): boolean {
  return getConfig().cloud || getConfig().enterprise;
}

export function hasMfaSupport(): boolean {
  return isPremiumFeature();
}
