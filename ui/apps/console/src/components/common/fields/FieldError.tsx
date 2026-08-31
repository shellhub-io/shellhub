import { ReactNode } from "react";

/**
 * A field's error message. It carries the id the input points at through aria-describedby, which
 * is what makes the error read out with the field rather than in isolation.
 */
export default function FieldError({
  id,
  role,
  children,
}: {
  id: string;
  role?: "alert" | "status";
  children: ReactNode;
}) {
  if (!children) return null;

  return (
    <p id={id} role={role} className="text-2xs text-accent-red mt-1.5">
      {children}
    </p>
  );
}
