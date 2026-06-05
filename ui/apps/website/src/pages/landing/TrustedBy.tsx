export function TrustedBy() {
  return (
    <section className="py-10 px-6 border-t border-border">
      <p className="text-center text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-secondary mb-6">Trusted by teams across the globe</p>
      <div className="flex items-center justify-center gap-10 md:gap-14 flex-wrap opacity-40">
        {["TechCorp", "CloudBase", "DevOps Inc", "NetSecure", "EdgeStack", "IoTWorks"].map((name) => (
          <span key={name} className="text-lg font-semibold tracking-tight whitespace-nowrap">{name}</span>
        ))}
      </div>
    </section>
  );
}
