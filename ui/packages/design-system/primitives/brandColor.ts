import { C } from "../constants";

/**
 * Which of the two fixed fills a brand mark uses: the brand violet, or white for placing the
 * mark on a coloured or photographic background.
 */
export type BrandVariant = "primary" | "inverted";

/** Brand marks render in two fixed colors only — never themed or inherited. */
export const brandFill = (variant: BrandVariant): string =>
  variant === "inverted" ? "#FFFFFF" : C.primary;
