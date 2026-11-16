import CheckboxField from "@/components/common/fields/CheckboxField";
import OptionCard from "./OptionCard";

/**
 * ReplaceOfflineDevicesField toggles whether a device enrolling with the key may replace an
 * offline device that reports the same identity, which is how a reflashed device gets its name
 * back. Unchecked, such a device stays pending.
 */
export default function ReplaceOfflineDevicesField({
  id,
  checked,
  onChange,
}: {
  id: string;
  checked: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <OptionCard active={checked}>
      <CheckboxField
        id={id}
        label="Replace offline devices"
        description="A device that reports the same identity as an offline device takes its name and sessions, and ShellHub deletes the offline device's record. Unchecked, the new device stays pending. Turn it off for a key you can't keep secret."
        checked={checked}
        onChange={onChange}
      />
    </OptionCard>
  );
}
