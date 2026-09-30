import type { ReactNode } from "react";

interface ScreenIntroProps {
  eyebrow: string;
  title: string;
  lead?: ReactNode;
}

/**
 * The heading every screen outside the console opens with: the small mono eyebrow naming the
 * flow, the title and the line under it, so the sign-in screens, the guards and the first run
 * read as one family.
 */
export default function ScreenIntro({
  eyebrow,
  title,
  lead,
}: ScreenIntroProps) {
  return (
    <>
      <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary mb-2">
        {eyebrow}
      </p>
      <h1 className="text-2xl font-bold tracking-tight text-text-primary mb-1.5">
        {title}
      </h1>
      {lead && <p className="text-sm text-text-secondary mb-8">{lead}</p>}
    </>
  );
}
