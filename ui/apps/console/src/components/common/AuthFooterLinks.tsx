import { getConfig } from "@/env";
import { BookOpenIcon, UserPlusIcon } from "@heroicons/react/24/outline";
import { GithubIcon } from "@shellhub/design-system/primitives";
import { Link } from "react-router-dom";

export default function AuthFooterLinks() {
  return (
    <div
      className="flex items-center justify-center gap-6 mt-10 animate-fade-in"
      style={{ animationDelay: "800ms" }}
    >
      <a
        href="https://docs.shellhub.io"
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
      >
        <BookOpenIcon className="w-3.5 h-3.5" />
        Documentation
      </a>
      <span className="w-px h-3 bg-border" />
      <a
        href="https://github.com/shellhub-io/shellhub"
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
      >
        <GithubIcon className="w-3.5 h-3.5" />
        Community
      </a>
      {getConfig().cloud && (
        <>
          <span className="w-px h-3 bg-border" />
          <Link
            to="/sign-up"
            className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <UserPlusIcon className="w-3.5 h-3.5" />
            Sign Up
          </Link>
        </>
      )}
    </div>
  );
}
