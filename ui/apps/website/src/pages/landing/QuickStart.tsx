import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { CommandBlock, Section, SectionHeader } from "@/components";
import { Reveal } from "@shellhub/design-system/components";
import { docsUrl, signupUrl } from "@/links";

const SETUP_CMD = [
  "git clone https://github.com/shellhub-io/shellhub.git",
  "cd shellhub",
  "make start",
].join("\n");

const linkClass =
  "inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:gap-2.5 transition-all group";

/**
 * Landing section with the self-hosting commands, anchored as #self-host for the hero's link,
 * and the Cloud alternative for a reader who would rather not run the server.
 */
export function QuickStart() {
  return (
    <Section id="self-host" bordered={false} className="scroll-mt-20">
      <SectionHeader
        className="mb-10"
        eyebrow="Self-host"
        title="Run your own server."
        subtitle="Docker, Docker Compose and openssl are all it takes. make start generates the keys and brings the stack up."
        subtitleClassName="mt-3 max-w-lg"
      />

      <Reveal>
        <CommandBlock command={SETUP_CMD} className="max-w-xl mx-auto" />
      </Reveal>

      <Reveal className="flex flex-wrap justify-center gap-x-8 gap-y-3 mt-6">
        <a
          href={`${docsUrl}/selfhosted/quickstart`}
          target="_blank"
          rel="noopener noreferrer"
          className={linkClass}
        >
          Self-hosting guide
          <ArrowRightIcon
            className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform"
            aria-hidden="true"
            strokeWidth={2.5}
          />
        </a>
        <a href={signupUrl} className={linkClass}>
          Rather not run it? Use Cloud
          <ArrowRightIcon
            className="w-3.5 h-3.5 group-hover:translate-x-0.5 transition-transform"
            aria-hidden="true"
            strokeWidth={2.5}
          />
        </a>
      </Reveal>
    </Section>
  );
}
