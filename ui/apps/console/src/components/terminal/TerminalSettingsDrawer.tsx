import { useState } from "react";
import Drawer from "../common/Drawer";
import TerminalSettingsBody from "./TerminalSettingsBody";

interface Props {
  open: boolean;
  onClose: () => void;
}

/**
 * The terminal's appearance settings, and how the session player shows its controls. They apply
 * to every open terminal and player at once, as a preference rather than per-session state.
 */
export default function TerminalSettingsDrawer({ open, onClose }: Props) {
  const [hasOpened, setHasOpened] = useState(open);
  if (open && !hasOpened) setHasOpened(true);
  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Terminal Settings"
    >
      {hasOpened && <TerminalSettingsBody />}
    </Drawer>
  );
}
