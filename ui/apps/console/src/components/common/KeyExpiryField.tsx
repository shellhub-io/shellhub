import RadioGroupField from "@/components/common/fields/RadioGroupField";
import RadioPill from "@/components/common/fields/RadioPill";
import { EXPIRY_OPTIONS } from "@/pages/team/helpers";

/**
 * How long an enrolled key keeps working. Offered wherever a key is enrolled —
 * the add-a-key drawer, the service-account drawer, and the approval screen a
 * held login sends you to — so the choice reads the same everywhere.
 *
 * expiresIn is an EXPIRY_OPTIONS value as a string ("-1" = never), which is the
 * default: a key nobody gave a deadline to should outlive the session that
 * created it.
 */
export default function KeyExpiryField({
  expiresIn,
  onExpiresInChange,
  label = "Expiration",
}: {
  expiresIn: string;
  onExpiresInChange: (value: string) => void;
  label?: string;
}) {
  return (
    <RadioGroupField
      label={label}
      value={expiresIn}
      onChange={onExpiresInChange}
      containerClassName="flex flex-wrap gap-1.5"
    >
      {EXPIRY_OPTIONS.map((opt) => (
        <RadioPill
          key={opt.value}
          value={String(opt.value)}
          label={opt.label}
        />
      ))}
    </RadioGroupField>
  );
}
