/**
 * The column a screen outside the console sits in: the form's width, or the wide one for a
 * screen that shows more than a form.
 */
export const columnWidths = {
  md: "max-w-md",
  "2xl": "max-w-2xl",
} as const;

/**
 * Which column a screen outside the console asks for.
 */
export type ColumnWidth = keyof typeof columnWidths;
