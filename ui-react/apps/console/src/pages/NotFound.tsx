import { Link } from "react-router-dom";
import { ArrowLeftIcon } from "@heroicons/react/24/outline";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="text-center animate-fade-in">
        <p className="text-7xl font-mono font-bold text-border mb-2">404</p>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-6">
          Page not found
        </p>
        <p className="text-sm text-text-secondary mb-8 max-w-xs mx-auto">
          The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-2 text-xs font-medium text-primary hover:text-primary-400 transition-colors"
        >
          <ArrowLeftIcon className="w-3.5 h-3.5" strokeWidth={2} />
          Back to Dashboard
        </Link>
      </div>
    </div>
  );
}
