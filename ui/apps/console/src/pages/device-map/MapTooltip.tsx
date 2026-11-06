import { compactCount } from "./mapModel";

/**
 * What the pointer is over on the map: a point or cluster, where it sits in the map's pixels, and
 * its counts.
 */
export interface MapHover {
  x: number;
  y: number;
  devices: number;
  online: number;
  cluster: boolean;
}

/**
 * The counts of the point or cluster under the pointer, drawn above it.
 */
export default function MapTooltip({ hover }: { hover: MapHover }) {
  return (
    <div
      className="pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-full rounded-md border border-border bg-surface/95 px-2.5 py-1.5 text-xs shadow-float"
      style={{ left: hover.x, top: hover.y - 14 }}
    >
      <p className="font-semibold text-text-primary">
        {compactCount(hover.devices)}{" "}
        {hover.devices === 1 ? "device" : "devices"}
        {hover.cluster && (
          <span className="font-normal text-text-muted"> · cluster</span>
        )}
      </p>
      <p className="font-mono text-2xs text-text-secondary">
        <span className="text-accent-green">
          {compactCount(hover.online)} online
        </span>
        {" · "}
        {compactCount(hover.devices - hover.online)} offline
      </p>
    </div>
  );
}
