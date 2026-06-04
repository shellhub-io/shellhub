import { ReactNode } from "react";

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
