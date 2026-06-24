import { CommandBlock, Section, SectionHeader } from "@/components/marketing";
import { Reveal } from "./components";
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
          <svg
            className="w-3 h-3 group-hover:translate-x-0.5 transition-transform"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2.5}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"
            />
          </svg>
        </a>
      </Reveal>
    </Section>
  );
}
