interface OnlineDotProps {
  online?: boolean;
}

export default function OnlineDot({ online }: OnlineDotProps) {
  if (online) {
    return (
      <span
        className="relative flex h-2.5 w-2.5 mx-auto"
        aria-label="Online"
        role="img"
      >
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-green opacity-40" />
        <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]" />
      </span>
    );
  }
  return (
    <span
      className="block w-2.5 h-2.5 rounded-full mx-auto bg-text-muted/30"
      aria-label="Offline"
      role="img"
    />
  );
}
