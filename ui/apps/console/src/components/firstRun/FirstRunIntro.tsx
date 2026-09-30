/**
 * The heading a first-run screen opens with: the eyebrow greeting, the title and the line under
 * it. Without a title it is the trail's own heading.
 */
export default function FirstRunIntro({
  eyebrow,
  title = "Get your first shell",
  lead = "From nothing to a shell on one of your devices. This page follows along, and picks up where you left off if you leave.",
}: {
  eyebrow: string;
  title?: string;
  lead?: string;
}) {
  return (
    <>
      <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary mb-2">
        {eyebrow}
      </p>
      <h1 className="text-2xl font-bold tracking-tight text-text-primary mb-1.5">
        {title}
      </h1>
      <p className="text-sm text-text-secondary mb-8">{lead}</p>
    </>
  );
}
