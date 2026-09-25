import { create } from "zustand";
import { persist } from "zustand/middleware";

/**
 * The ways the session player can show its floating controls, in the order the settings list them
 * and the H key steps through them. auto fades them while a recording plays and nothing is
 * happening; always keeps them up; hidden leaves only the keyboard shortcuts. notice is what the
 * player says when H lands on one.
 */
export const PLAYER_CONTROLS = [
  { value: "auto", label: "Auto-hide", notice: "Controls: auto-hide" },
  { value: "always", label: "Always", notice: "Controls: always shown" },
  { value: "hidden", label: "Hidden", notice: "Controls: hidden" },
] as const;

/**
 * One of the session player's controls modes. Derived from PLAYER_CONTROLS, so the modes, the
 * settings tiles and the H cycle cannot drift apart.
 */
export type PlayerControls = (typeof PLAYER_CONTROLS)[number]["value"];

/**
 * The mode H moves to from controls, wrapping from the last back to the first.
 */
export function nextPlayerControls(controls: PlayerControls) {
  const at = PLAYER_CONTROLS.findIndex((mode) => mode.value === controls);
  return PLAYER_CONTROLS[(at + 1) % PLAYER_CONTROLS.length];
}

interface SessionPlayerState {
  controls: PlayerControls;
  setControls: (controls: PlayerControls) => void;
}

/**
 * The session player's controls preference, remembered across reloads.
 */
export const useSessionPlayerStore = create<SessionPlayerState>()(
  persist(
    (set) => ({
      controls: "auto",
      setControls: (controls) => set({ controls }),
    }),
    {
      name: "playerControls",
      partialize: (state) => ({ controls: state.controls }),
    },
  ),
);
