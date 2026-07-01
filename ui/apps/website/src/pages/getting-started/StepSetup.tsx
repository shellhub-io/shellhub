import { ArrowLeftIcon, ArrowRightIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { CommandBlock } from "@/components/marketing";
import { docsUrl } from "@/links";
import { Reveal } from "../landing/components";

const DOCKER_CMD = "docker run -d -p 80:80 shellhubio/shellhub";

interface StepSetupProps {
  onBack: () => void;
}

export function StepSetup({ onBack }: StepSetupProps) {
  return (
    <div className="max-w-xl mx-auto w-full">
      <Reveal>
        <CommandBlock command={DOCKER_CMD} className="mb-6" />
      </Reveal>

      <Reveal delay={0.1}>
        <div className="space-y-3 mb-8">
          <p className="text-sm text-text-secondary leading-relaxed">
            This starts ShellHub on port 80. Open{" "}
            <code className="font-mono text-xs bg-surface px-1.5 py-0.5 rounded border border-border">
              http://localhost
            </code>{" "}
            in your browser and create your account.
          </p>
          <p className="text-sm text-text-secondary leading-relaxed">
            Then install the ShellHub agent on each device you want to manage.
          </p>
        </div>
      </Reveal>

      <Reveal delay={0.15}>
        <div className="flex items-center justify-between">
          <Button
            variant="secondary"
            className="group"
            onClick={onBack}
            icon={
              <ArrowLeftIcon
                className="w-4 h-4 group-hover:-translate-x-0.5 transition-transform duration-300"
                aria-hidden="true"
              />
            }
          >
            Back
          </Button>

          <a
            href={`${docsUrl}/getting-started`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs font-medium text-primary hover:gap-2.5 transition-all group"
          >
            Full documentation
            <ArrowRightIcon
              className="w-3 h-3 group-hover:translate-x-0.5 transition-transform"
              aria-hidden="true"
            />
          </a>
        </div>
      </Reveal>
    </div>
  );
}
