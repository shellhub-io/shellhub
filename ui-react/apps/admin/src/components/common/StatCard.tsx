import { ReactNode } from "react";
import { Link } from "react-router-dom";

interface StatCardProps {
  icon: ReactNode;
  title: string;
  value: number | string;
  linkLabel: string;
  linkTo: string;
  accent?: string;
}

export default function StatCard({
  icon,
  title,
  value,
  linkLabel,
  linkTo,
  accent,
}: StatCardProps) {
  return (
    <div className="bg-card border border-border rounded-lg p-6 flex flex-col items-center text-center group hover:border-primary/30 transition-all duration-300">
      <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary mb-5">
        {icon}
      </div>

      <p className="text-2xs font-mono font-medium uppercase tracking-label text-text-muted mb-2">
        {title}
      </p>

      <p
        className={`text-4xl font-mono font-bold mb-5 tabular-nums ${accent ?? "text-text-primary"}`}
      >
        {value}
      </p>

      <Link
        to={linkTo}
        className="text-xs font-medium text-primary hover:text-primary-400 transition-colors"
      >
        {linkLabel} &rarr;
      </Link>
    </div>
  );
}
