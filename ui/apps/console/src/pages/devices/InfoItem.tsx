import CopyButton from "@/components/common/CopyButton";

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";

interface InfoItemProps {
  label: string;
  value: string;
  mono?: boolean;
  copyable?: boolean;
  truncate?: number;
}

export default function InfoItem({
  label,
  value,
  mono,
  copyable,
  truncate,
}: InfoItemProps) {
  const display = truncate && value ? value.slice(0, truncate) : value;

  return (
    <div>
      <dt className={LABEL}>{label}</dt>
      <dd className="flex items-center gap-1 mt-0.5">
        <span
          className={`text-sm text-text-primary ${mono ? "font-mono text-xs" : "font-medium"}`}
        >
          {display || "—"}
        </span>
        {copyable && value && <CopyButton text={value} />}
      </dd>
    </div>
  );
}
