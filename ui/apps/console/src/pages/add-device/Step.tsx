import type { ReactNode } from "react";
import { LABEL_BASE } from "@/utils/styles";

interface StepProps {
  n: number;
  label: ReactNode;
  labelId?: string;
  children: ReactNode;
}

/** One numbered step of adding a device: its number and label, then what the step asks for. */
export default function Step({ n, label, labelId, children }: StepProps) {
  return (
    <div>
      <div className="flex items-center gap-2.5 mb-3">
        <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
          {n}
        </span>
        <span id={labelId} className={LABEL_BASE}>
          {label}
        </span>
      </div>
      <div className="space-y-3">{children}</div>
    </div>
  );
}
