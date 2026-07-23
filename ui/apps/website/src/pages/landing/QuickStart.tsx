import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { CommandBlock, Section, SectionHeader } from "@/components";
import { Reveal } from "@shellhub/design-system/components";
import { docsUrl } from "@/links";

const DOCKER_CMD = "docker run -d -p 80:80 shellhubio/shellhub";

export function QuickStart() {
  return (
    <Section bordered={false}>
      <SectionHeader
        className="mb-10"
        eyebrow="Quick Start"
        title="Try ShellHub in seconds."
        subtitle="Get up and running with a single command."
        subtitleClassName="mt-3 max-w-md"
      />

      <Reveal>
        <CommandBlock command={DOCKER_CMD} className="max-w-xl mx-auto" />
      </Reveal>

      <Reveal className="text-center mt-5">
        <a
          href={docsUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 text-xs font-medium text-primary hover:gap-2.5 transition-all group"
        >
          Full installation guide
          <ArrowRightIcon
            className="w-3 h-3 group-hover:translate-x-0.5 transition-transform"
            aria-hidden="true"
            strokeWidth={2.5}
          />
        </a>
      </Reveal>
    </Section>
  );
}
