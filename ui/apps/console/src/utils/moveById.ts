/**
 * The list with the item whose id is given moved to sit at index to, counted after its removal.
 * An unknown id, an index out of range or a move to where the item already is returns the list
 * itself, so a caller can tell nothing moved by identity; the input is never changed.
 */
export function moveById<T extends { id: string }>(
  list: T[],
  id: string,
  to: number,
): T[] {
  const from = list.findIndex((item) => item.id === id);
  if (from < 0 || to < 0 || to >= list.length || to === from) return list;
  const next = [...list];
  const [item] = next.splice(from, 1);
  next.splice(to, 0, item);
  return next;
}
