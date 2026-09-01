/**
 * Shots are addressed by id alone, so the id is the whole filename. A flat
 * directory is what lets the tag and the capture tool agree on a path without
 * either of them carrying a second list.
 */
export const SHOTS_DIR = "img/shots";

/** The URL a page requests for a shot. */
export function shotImageHref(id: string): string {
  return `/${SHOTS_DIR}/${id}.png`;
}

/** The shot's path inside `public/`. */
export function shotImageFile(id: string): string {
  return `${SHOTS_DIR}/${id}.png`;
}
