import { formatExpiry } from "./date";
import type { GetLicenseResponse } from "../client/types.gen";

export function formatLicenseTimestamp(value: number): string {
  if (value === -1) return "Now";
  return formatExpiry(value);
}

export function formatDeviceCount(value: number): string {
  return value === -1 ? "Unlimited" : String(value);
}

export function formatRegions(regions: string[]): string {
  return regions.length === 0 ? "Global" : regions.join(", ");
}

type Features = GetLicenseResponse["features"];

type DisplayFeature
  = { name: string; label: string; type: "boolean"; value: boolean }
    | { name: string; label: string; type: "number"; value: number };

// "login_link" and "reports" are excluded to match the Vue admin UI.
export function getDisplayFeatures(features: Features): DisplayFeature[] {
  return [
    { name: "devices", label: "Devices", type: "number", value: features.devices },
    { name: "session_recording", label: "Session recording", type: "boolean", value: features.session_recording },
    { name: "firewall_rules", label: "Firewall rules", type: "boolean", value: features.firewall_rules },
    { name: "billing", label: "Billing", type: "boolean", value: features.billing },
  ];
}

export function validateLicenseFile(file: File): string | null {
  if (!file.name.toLowerCase().endsWith(".dat")) return "Only .dat files are allowed";
  if (file.size >= 32 * 1024) return "File must be smaller than 32 KB";
  return null;
}

interface LicenseFlags {
  expired: boolean;
  about_to_expire: boolean;
  grace_period: boolean;
}

interface AlertConfig {
  variant: "info" | "warning" | "error";
  message: string;
}

export function getLicenseAlertConfig(license: LicenseFlags | null): AlertConfig | null {
  if (!license) {
    return { variant: "info", message: "You do not have an installed license" };
  }
  if (license.expired && !license.grace_period) {
    return { variant: "error", message: "Your license has expired!" };
  }
  if (license.expired && license.grace_period) {
    return { variant: "warning", message: "Your license has expired, but you are still within the grace period." };
  }
  if (license.about_to_expire) {
    return { variant: "info", message: "Your license is about to expire!" };
  }
  return null;
}
