// primitives/index.ts — public surface of the design-system primitives.
// cn is intentionally NOT exported from this component barrel. Within the package, import it
// from "./cn". External consumers should import it via the "@shellhub/design-system/cn" subpath.

export { Button } from "./Button";
export type { ButtonVariant, ButtonSize } from "./Button";
export { IconButton } from "./IconButton";
export type { IconButtonVariant, IconButtonSize } from "./IconButton";
export { Badge } from "./Badge";
export type { BadgeColor, BadgeShape, BadgeProps } from "./Badge";
export { Card } from "./Card";
export { IconBadge } from "./IconBadge";
export type { Palette, IconBadgeSize, IconBadgeProps } from "./IconBadge";
export { Spinner } from "./Spinner";
export type { SpinnerSize, SpinnerTone } from "./Spinner";
export { StatusDot } from "./StatusDot";
export type {
  StatusDotColor,
  StatusDotSize,
  StatusDotProps,
} from "./StatusDot";
export { WindowChrome } from "./WindowChrome";
export type {
  WindowChromeVariant,
  WindowChromeSize,
  WindowChromeAccent,
  WindowChromeProps,
} from "./WindowChrome";
export { Callout } from "./Callout";
export type { CalloutVariant, CalloutProps } from "./Callout";
