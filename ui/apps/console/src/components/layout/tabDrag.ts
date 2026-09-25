/**
 * Where a tab of a group sits, measured when a drag starts, left to right.
 */
export interface TabSlot {
  id: string;
  left: number;
  width: number;
}

/**
 * A tab being dragged along its group: the slots as they were when the drag began, and how far
 * the pointer has carried the tab since.
 */
export interface TabDrag {
  id: string;
  slots: TabSlot[];
  offset: number;
}

function gapOf(slots: TabSlot[]): number {
  const [first, second] = slots;
  return first && second ? second.left - (first.left + first.width) : 0;
}

/**
 * How far the dragged tab is drawn from its slot: the pointer's offset, held inside the group so
 * the tab cannot leave it.
 */
export function draggedOffset(drag: TabDrag): number {
  const slot = drag.slots.find((s) => s.id === drag.id);
  const first = drag.slots[0];
  const last = drag.slots[drag.slots.length - 1];
  if (!slot || !first || !last) return 0;
  const min = first.left - slot.left;
  const max = last.left + last.width - (slot.left + slot.width);
  return Math.min(max, Math.max(min, drag.offset));
}

/**
 * The index the dragged tab would take if dropped now. A tab to its right is passed once the
 * dragged tab's right edge crosses that tab's centre, and one to its left once its left edge does,
 * so the others give way as soon as the dragged tab overlaps half of them.
 */
export function dropIndex(drag: TabDrag): number {
  const from = drag.slots.findIndex((s) => s.id === drag.id);
  if (from < 0) return -1;
  const slot = drag.slots[from];
  const left = slot.left + draggedOffset(drag);
  const right = left + slot.width;
  return drag.slots.filter((s, i) => {
    if (i === from) return false;
    const centre = s.left + s.width / 2;
    return i < from ? left > centre : right > centre;
  }).length;
}

/**
 * How far another tab of the group slides to make room: a slot's width towards where the dragged
 * tab came from, for each tab between its old place and its new one, and nothing for the rest.
 */
export function makeRoomOffset(drag: TabDrag, id: string): number {
  const from = drag.slots.findIndex((s) => s.id === drag.id);
  const at = drag.slots.findIndex((s) => s.id === id);
  if (from < 0 || at < 0 || at === from) return 0;
  const to = dropIndex(drag);
  const room = drag.slots[from].width + gapOf(drag.slots);
  if (from < to && at > from && at <= to) return -room;
  if (from > to && at >= to && at < from) return room;
  return 0;
}
