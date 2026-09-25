import { useState, type PointerEvent } from "react";
import {
  draggedOffset,
  dropIndex,
  makeRoomOffset,
  type TabDrag,
} from "./tabDrag";

/**
 * What one tab needs to take part in reordering: where to draw it, whether it is the one being
 * dragged or one making room, and the handlers that drive the drag.
 */
export interface TabReorder {
  offset: number;
  dragging: boolean;
  settling: boolean;
  onPointerDown: (e: PointerEvent<HTMLElement>) => void;
  onPointerMove: (e: PointerEvent<HTMLElement>) => void;
  onPointerUp: () => void;
  onPointerCancel: () => void;
  onMove: (step: -1 | 1) => void;
}

const DRAG_THRESHOLD = 4;

function measure(ids: string[], strip: Element): TabDrag["slots"] {
  return ids.flatMap((tabId) => {
    const el = strip.querySelector(`[data-tab-id="${CSS.escape(tabId)}"]`);
    if (!el) return [];
    const box = el.getBoundingClientRect();
    return [{ id: tabId, left: box.left, width: box.width }];
  });
}

/**
 * Reorders tabs by dragging them within their group, or by a step from the keyboard. A press
 * becomes a drag once the pointer has moved a few pixels, so a click still selects. The tabs'
 * places are measured from the strip, found as the pressed tab's parent with each tab marked by
 * data-tab-id, when the drag starts. Releasing drops the tab where it was drawn, and cancel (or a
 * cancelled pointer) puts it back. orderOf gives a group's order as it would be if dropped now.
 */
export function useTabReorder() {
  const [press, setPress] = useState<{
    id: string;
    x: number;
    ids: string[];
    move: (id: string, to: number) => void;
  } | null>(null);
  const [drag, setDrag] = useState<TabDrag | null>(null);

  const endDrag = (commit: boolean) => {
    if (commit && drag && press && drag.id === press.id) {
      press.move(drag.id, dropIndex(drag));
    }
    setPress(null);
    setDrag(null);
  };

  const tab = (
    ids: string[],
    move: (id: string, to: number) => void,
    id: string,
  ): TabReorder => {
    const inGroup = drag !== null && ids.includes(drag.id);
    return {
      offset: !inGroup
        ? 0
        : drag.id === id
          ? draggedOffset(drag)
          : makeRoomOffset(drag, id),
      dragging: drag?.id === id,
      settling: inGroup,
      onPointerDown: (e) => {
        if (e.button !== 0) return;
        e.currentTarget.setPointerCapture(e.pointerId);
        setDrag(null);
        setPress({ id, x: e.clientX, ids, move });
      },
      onPointerMove: (e) => {
        if (!press || press.id !== id) return;
        const offset = e.clientX - press.x;
        if (drag) {
          setDrag({ ...drag, offset });
        } else if (Math.abs(offset) > DRAG_THRESHOLD) {
          const strip = e.currentTarget.parentElement;
          if (strip) setDrag({ id, slots: measure(press.ids, strip), offset });
        }
      },
      onPointerUp: () => endDrag(true),
      onPointerCancel: () => endDrag(false),
      onMove: (step) => move(id, ids.indexOf(id) + step),
    };
  };

  const orderOf = (ids: string[]) => {
    if (!drag || !ids.includes(drag.id)) return ids;
    const rest = ids.filter((tabId) => tabId !== drag.id);
    const at = dropIndex(drag);
    return [...rest.slice(0, at), drag.id, ...rest.slice(at)];
  };

  const cancel = () => {
    if (drag) endDrag(false);
  };

  return { tab, orderOf, cancel };
}
