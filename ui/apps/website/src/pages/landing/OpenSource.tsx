import { useEffect, useState } from "react";
import { StarIcon } from "@heroicons/react/24/solid";
import { GithubIcon } from "@shellhub/design-system/primitives";
import { ActionButton, Section, SectionHeader } from "@/components";
import { Reveal } from "@shellhub/design-system/components";
import { C } from "@shellhub/design-system/constants";
import { githubUrl, signupUrl } from "@/links";

function formatStars(count: number) {
  return `${(count / 1000).toFixed(1)}k`;
}

export function OpenSource() {
  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    fetch("https://api.github.com/repos/shellhub-io/shellhub")
      .then((r) => {
        if (!r.ok) throw new Error(r.statusText);
        return r.json() as Promise<{ stargazers_count: number }>;
      })
      .then((d) => setStars(d.stargazers_count))
      .catch(() => {});
  }, []);
  return (
    <Section background="surface" className="border-b border-border" centered>
      <SectionHeader
        className="mb-3"
        eyebrow="Open Source"
        title="Built in the open."
      />
      <Reveal>
        <p className="text-sm text-text-secondary max-w-md mx-auto mb-8 leading-relaxed">
          ShellHub is open source. Self-host it, customize it, or use our
          managed cloud.
        </p>
      </Reveal>
      <Reveal>
        {stars !== null && (
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md border border-border bg-card text-sm text-text-secondary mb-8">
            <StarIcon className="w-3.5 h-3.5" style={{ color: C.yellow }} />
            <span className="text-xs font-mono">
              {formatStars(stars)} stars on GitHub
            </span>
          </div>
        )}
      </Reveal>
      <Reveal>
        <div className="flex gap-3 justify-center flex-wrap">
          <ActionButton
            action={{
              label: "View on GitHub",
              href: githubUrl,
              external: true,
            }}
            variant="outline"
            size="lg"
            icon={<GithubIcon width={16} height={16} />}
          />
          <ActionButton
            action={{
              label: "Try Cloud Free",
              href: signupUrl,
              external: true,
            }}
            size="lg"
            glow={false}
            iconRight={null}
          />
        </div>
      </Reveal>
    </Section>
  );
}
