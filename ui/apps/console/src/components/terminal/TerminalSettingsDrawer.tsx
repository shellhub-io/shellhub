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
  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Terminal Settings"
      width="sm"
      bodyClassName="flex-1 overflow-y-auto"
    >
      <TerminalSettingsBody />
    </Drawer>
  );
}
