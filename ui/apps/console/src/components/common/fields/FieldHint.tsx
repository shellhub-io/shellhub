import { ReactNode } from "react";

/**
 * A field's hint, associated with the input the same way an error is.
 */
export default function FieldHint({
  id,
  children,
}: {
  id: string;
  children: ReactNode;
}) {
  if (!children) return null;

  return (
    <p id={id} className="text-2xs text-text-muted mt-1.5">
      {children}
    </p>
  );
}
