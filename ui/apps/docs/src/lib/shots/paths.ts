/**
 * Shots are addressed by id alone, so the id is the whole filename. A flat
 * directory is what lets the tag and the capture tool agree on a path without
 * either of them carrying a second list.
 */
export const SHOTS_DIR = "img/shots";

export function shotImageHref(id: string): string {
  return `/${SHOTS_DIR}/${id}.png`;
}

export function shotImageFile(id: string): string {
  return `${SHOTS_DIR}/${id}.png`;
}
