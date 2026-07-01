import { StarIcon } from "@heroicons/react/24/solid";
import { Button, GithubIcon } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal } from "./components";
import { C } from "./constants";
import { githubUrl, signupUrl } from "@/links";

export function OpenSource() {
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
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md border border-border bg-card text-sm text-text-secondary mb-8">
          <StarIcon className="w-3.5 h-3.5" style={{ color: C.yellow }} />
          <span className="text-xs font-mono">3,200+ stars on GitHub</span>
        </div>
      </Reveal>
      <Reveal>
        <div className="flex gap-3 justify-center flex-wrap">
          <Button
            as="a"
            href={githubUrl}
            target="_blank"
            rel="noopener noreferrer"
            variant="outline"
            size="lg"
            icon={<GithubIcon width={16} height={16} />}
          >
            View on GitHub
          </Button>
          <Button
            as="a"
            href={signupUrl}
            target="_blank"
            rel="noopener noreferrer"
            variant="primary"
            size="lg"
          >
            Try Cloud Free
          </Button>
        </div>
      </Reveal>
    </Section>
  );
}
