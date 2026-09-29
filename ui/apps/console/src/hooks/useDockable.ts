import {
  useEffect,
  useRef,
  useState,
  type PointerEvent,
  type RefObject,
  type TransitionEvent,
} from "react";

const FLOAT_INSET = 8;
const UNDOCK_DISTANCE = 12;
const DOCK_REACH = 56;
const EDGE_ZONE = 64;
const EDGE_SPEED = 14;

type Placement = "docked" | "floating" | "docking";

function useEdgeScroll(
  active: boolean,
  slot: RefObject<HTMLDivElement | null>,
  float: RefObject<{ top: number; height: number }>,
  onScrolled: (overDock: boolean) => void,
) {
  useEffect(() => {
    if (!active) return;
    const page = slot.current?.closest("main");
    if (!page) return;
    let id = 0;
    const step = () => {
      const bounds = page.getBoundingClientRect();
      const { top, height } = float.current;
      const up = bounds.top + EDGE_ZONE - top;
      const down = top + height - (bounds.bottom - EDGE_ZONE);
      const by =
        up > 0
          ? -Math.ceil((up / EDGE_ZONE) * EDGE_SPEED)
          : down > 0
            ? Math.ceil((down / EDGE_ZONE) * EDGE_SPEED)
            : 0;
      if (by !== 0) {
        page.scrollBy(0, by);
        const target = slot.current?.getBoundingClientRect().top ?? top;
        onScrolled(Math.abs(top - target) < DOCK_REACH);
      }
      id = requestAnimationFrame(step);
    };
    id = requestAnimationFrame(step);
    return () => cancelAnimationFrame(id);
  }, [active, slot, float, onScrolled]);
}

/**
 * Lets an element be dragged out of its place in the page to float fixed over it, and back in.
 * Dragged past a few pixels the element undocks; slot is the placeholder left behind, which the
 * caller renders while placement is not "docked" and opens (slotOpen) while the element hovers
 * near it. Released near the slot the element glides into it (placement "docking") and docks on
 * the transition's end, or at once under reduced motion; released elsewhere it keeps floating.
 * Dragged near the edges of the page's scroll container, AppLayout's main, the page scrolls.
 * When enabled turns false everything goes back to docked, so a layout that stops offering the
 * drag never keeps a floating element.
 */
export function useDockable(enabled: boolean) {
  const slot = useRef<HTMLDivElement>(null);
  const start = useRef<{
    pointer: number;
    top: number;
    left: number;
    width: number;
  } | null>(null);
  const float = useRef({ top: 0, height: 0 });
  const [placement, setPlacement] = useState<Placement>("docked");
  const [dragging, setDragging] = useState(false);
  const [overDock, setOverDock] = useState(false);
  const [justUndocked, setJustUndocked] = useState(false);
  const [frame, setFrame] = useState({ top: 0, left: 0, width: 0 });
  const [wasEnabled, setWasEnabled] = useState(enabled);
  if (enabled !== wasEnabled) {
    setWasEnabled(enabled);
    if (!enabled) {
      setPlacement("docked");
      setDragging(false);
      setOverDock(false);
      setJustUndocked(false);
    }
  }

  useEdgeScroll(dragging && placement === "floating", slot, float, setOverDock);

  const dockTop = () =>
    slot.current?.getBoundingClientRect().top ?? start.current?.top ?? 0;

  const dock = () => {
    setPlacement("docked");
    setJustUndocked(false);
    setOverDock(false);
  };

  const release = () => {
    if (!start.current) return;
    start.current = null;
    setDragging(false);
    if (placement !== "floating" || !overDock) {
      setOverDock(false);
      return;
    }
    const target = dockTop();
    const still = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (still || Math.abs(float.current.top - target) < 1) {
      dock();
      return;
    }
    setFrame((f) => ({ ...f, top: target }));
    setPlacement("docking");
  };

  return {
    slot,
    placement,
    dragging,
    overDock,
    slotOpen: justUndocked || overDock,
    frame,
    onDocked: (e: TransitionEvent<HTMLElement>) => {
      if (placement === "docking" && e.propertyName === "top") dock();
    },
    onDockInterrupted: () => {
      if (placement === "docking") dock();
    },
    handlers: {
      onPointerDown: (e: PointerEvent<HTMLElement>) => {
        if (!enabled || placement === "docking") return;
        const box = e.currentTarget.getBoundingClientRect();
        start.current = {
          pointer: e.clientY,
          top: box.top,
          left: box.left,
          width: box.width,
        };
        setDragging(true);
        e.currentTarget.setPointerCapture(e.pointerId);
      },
      onPointerMove: (e: PointerEvent<HTMLElement>) => {
        const from = start.current;
        if (!enabled || !from) return;
        const moved = e.clientY - from.pointer;
        if (placement === "docked" && Math.abs(moved) < UNDOCK_DISTANCE) return;
        const height = e.currentTarget.offsetHeight;
        const room = window.innerHeight - height - FLOAT_INSET;
        const top = Math.min(
          Math.max(FLOAT_INSET, from.top + moved),
          Math.max(FLOAT_INSET, room),
        );
        float.current = { top, height };
        if (placement === "docked") {
          setFrame({ top, left: from.left, width: from.width });
          setJustUndocked(true);
          setPlacement("floating");
        } else {
          setFrame((f) => ({ ...f, top }));
          setJustUndocked(false);
        }
        setOverDock(Math.abs(top - dockTop()) < DOCK_REACH);
      },
      onPointerUp: release,
      onPointerCancel: release,
    },
  };
}
