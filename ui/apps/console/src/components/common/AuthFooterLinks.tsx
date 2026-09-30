import { getConfig } from "@/env";
import { BookOpenIcon, TagIcon } from "@heroicons/react/24/outline";
import { GithubIcon } from "@shellhub/design-system/primitives";

/**
 * The line under every public screen of the community edition: the docs, the running version,
 * and the community, in three equal parts drawn the same way.
 */
export default function AuthFooterLinks() {
  const version = getConfig().version;
  const itemClass =
    "inline-flex items-center justify-center gap-1.5 text-xs text-text-muted";
  const linkClass = `${itemClass} hover:text-text-secondary transition-colors`;

  return (
    <div className="grid grid-cols-3 items-center mt-5 px-2">
      <a
        href="https://docs.shellhub.io"
        target="_blank"
        rel="noopener noreferrer"
        className={linkClass}
      >
        <BookOpenIcon className="w-3.5 h-3.5" />
        Documentation
      </a>
      <span className={itemClass}>
        <TagIcon className="w-3.5 h-3.5" />
        <span className="font-mono">{version}</span>
      </span>
      <a
        href="https://github.com/shellhub-io/shellhub"
        target="_blank"
        rel="noopener noreferrer"
        className={linkClass}
      >
        <GithubIcon className="w-3.5 h-3.5" />
        Community
      </a>
    </div>
  );
}
