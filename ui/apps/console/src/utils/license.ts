import { formatExpiry } from "./date";
import type { GetLicenseResponse } from "../client";

/**
 * Formats a license timestamp. -1 is the sentinel for a boundary already passed, and renders as
 * "Now" rather than as a date in 1969.
 */
export function formatLicenseTimestamp(value: number): string {
  if (value === -1) return "Now";
  return formatExpiry(value);
}

/**
 * Formats a licensed device count. -1 is the sentinel for no limit.
 */
export function formatDeviceCount(value: number): string {
  return value === -1 ? "Unlimited" : String(value);
}

/**
 * Formats the licensed regions. An empty list is not a licence with no regions but one with no
 * restriction, so it reads as "Global".
 */
export function formatRegions(regions: string[]): string {
  return regions.length === 0 ? "Global" : regions.join(", ");
}

type Features = GetLicenseResponse["features"];

type DisplayFeature
  = { name: string; label: string; type: "boolean"; value: boolean }
  | { name: string; label: string; type: "number"; value: number };

/**
 * The licence features as rows for the admin table, in display order. login_link and reports are
 * deliberately absent: the Vue admin UI does not show them, and the two views are kept in step
 * until it is retired.
 */
export function getDisplayFeatures(features: Features): DisplayFeature[] {
  return [
    { name: "devices", label: "Devices", type: "number", value: features.devices },
    { name: "session_recording", label: "Session recording", type: "boolean", value: features.session_recording },
    { name: "firewall_rules", label: "Firewall rules", type: "boolean", value: features.firewall_rules },
    { name: "billing", label: "Billing", type: "boolean", value: features.billing },
  ];
}

/**
 * Checks a licence file before upload, returning the reason to show or null when it is fine.
 * Both limits mirror the server, so an obvious mistake is caught without a round trip.
 */
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

/**
 * Which banner the licence state warrants, or null when there is nothing to say. Expiry and
 * grace period are separate flags, so an expired licence still inside its grace period warns
 * rather than errors — the instance is still working, and the message must not say otherwise.
 */
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
