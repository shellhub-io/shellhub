// primitives/index.ts — public surface of the design-system primitives.
// cn is intentionally NOT exported here; import it directly from "./cn" within the package.

export { Badge } from "./Badge";
export type { BadgeColor, BadgeShape, BadgeProps } from "./Badge";
export { Card } from "./Card";
export { IconBadge } from "./IconBadge";
export type { Palette, IconBadgeSize, IconBadgeProps } from "./IconBadge";
export { StatusDot } from "./StatusDot";
export type {
  StatusDotColor,
  StatusDotSize,
  StatusDotProps,
} from "./StatusDot";
