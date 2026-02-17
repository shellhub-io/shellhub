export function ShimmerCard({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`group relative ${className}`}>
      <div className="shimmer absolute inset-0 rounded-xl overflow-hidden opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" />
      <div className="relative">{children}</div>
    </div>
  );
}
