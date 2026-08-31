/** Page size every list page uses unless it sets its own. */
export const PER_PAGE = 10;

/**
 * How many pages a total spans. Rounds up, so a partial last page still counts.
 */
export function pageCount(totalCount: number, perPage: number = PER_PAGE): number {
  return Math.ceil(totalCount / perPage);
}
