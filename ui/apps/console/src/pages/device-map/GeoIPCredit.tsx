import type { GeoIPSource } from "@/types/geoip";

/**
 * The credit the GeoIP database that placed the devices requires, nothing while the source is
 * unknown.
 */
export default function GeoIPCredit({ source }: { source?: GeoIPSource }) {
  if (!source) return null;

  const credit =
    source === "maxmind" ? (
      <span>This product includes GeoLite Data created by MaxMind</span>
    ) : (
      <a
        href="https://db-ip.com"
        target="_blank"
        rel="noreferrer"
        className="hover:text-text-primary"
      >
        IP Geolocation by DB-IP
      </a>
    );

  return (
    <p className="absolute bottom-2 right-2 z-10 rounded bg-surface/80 px-2 py-0.5 text-3xs text-text-muted backdrop-blur">
      {credit}
    </p>
  );
}
