import { ReactNode } from "react";
import { LABEL, LABEL_BASE } from "@/utils/styles";

/**
 * A field label. hideLabel keeps it in the accessibility tree while hiding it visually, for a
 * field whose purpose is already clear from context — a search box, a row of one.
 */
export default function FieldLabel({
  htmlFor,
  id,
  hideLabel,
  adornment,
  children,
}: {
  htmlFor?: string;
  id?: string;
  hideLabel?: boolean;
  adornment?: ReactNode;
  children: ReactNode;
}) {
  if (adornment && !hideLabel) {
    return (
      <div className="flex items-center gap-2 mb-1.5">
        <label htmlFor={htmlFor} id={id} className={LABEL_BASE}>
          {children}
        </label>
        {adornment}
      </div>
    );
  }

  return (
    <label htmlFor={htmlFor} id={id} className={hideLabel ? "sr-only" : LABEL}>
      {children}
    </label>
  );
}
