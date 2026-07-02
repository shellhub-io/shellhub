import { C } from "../constants";

export type BrandVariant = "primary" | "inverted";

/** Brand marks render in two fixed colors only — never themed or inherited. */
export const brandFill = (variant: BrandVariant): string =>
  variant === "inverted" ? "#FFFFFF" : C.primary;
