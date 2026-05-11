import { ReactNode } from "react";

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
