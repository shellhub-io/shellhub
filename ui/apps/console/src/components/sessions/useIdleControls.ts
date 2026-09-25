import { useEffect, useRef, useState } from "react";

const IDLE_MS = 2000;

/**
 * When a player's floating controls should be up. wake counts as activity: it brings them up and
 * restarts the two-second idle wait. stow puts them away until the next wake, whatever else would
 * hold them. shown(pinned) says whether they are up, pinned being whatever the caller keeps them
 * up for regardless of activity, such as a paused recording. The pointer resting on the controls
 * also holds them, through the handlers in barProps.
 */
export function useIdleControls() {
  const [awake, setAwake] = useState(false);
  const [overBar, setOverBar] = useState(false);
  const [stowed, setStowed] = useState(false);
  const idleRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const stopWaiting = () => {
    if (idleRef.current !== null) clearTimeout(idleRef.current);
    idleRef.current = null;
  };

  useEffect(() => stopWaiting, []);

  const wake = () => {
    stopWaiting();
    setStowed(false);
    setAwake(true);
    idleRef.current = setTimeout(() => setAwake(false), IDLE_MS);
  };

  const stow = () => {
    stopWaiting();
    setAwake(false);
    setOverBar(false);
    setStowed(true);
  };

  return {
    wake,
    stow,
    shown: (pinned: boolean) => !stowed && (pinned || awake || overBar),
    barProps: {
      onPointerEnter: () => setOverBar(true),
      onPointerLeave: () => setOverBar(false),
    },
  };
}
