import type { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import CopyButton from "@/components/common/CopyButton";
import { LABEL_BASE } from "@/utils/styles";

interface InfoItemProps {
  label: string;
  value?: string;
  mono?: boolean;
  copyable?: boolean;
  truncate?: number;
  children?: ReactNode;
}

export default function InfoItem({
  label,
  value,
  mono,
  copyable,
  truncate,
  children,
}: InfoItemProps) {
  const display = truncate && value ? value.slice(0, truncate) : value;

  return (
    <div>
      <dt className={LABEL_BASE}>{label}</dt>
      <dd className="flex items-center gap-1 mt-0.5">
        {children ?? (
          <>
            <span
              className={cn(
                "text-sm text-text-primary",
                mono ? "font-mono text-xs" : "font-medium",
              )}
            >
              {display || "—"}
            </span>
            {copyable && value && <CopyButton text={value} />}
          </>
        )}
      </dd>
    </div>
  );
}
