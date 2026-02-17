import { Reveal, CopyBtn } from "../landing/components";

const DOCKER_CMD = "docker run -d -p 80:80 shellhubio/shellhub";

interface StepSetupProps {
  onBack: () => void;
}

export function StepSetup({ onBack }: StepSetupProps) {
  return (
    <div className="max-w-xl mx-auto w-full">
      <Reveal>
        <div className="bg-card border border-border rounded-xl overflow-hidden mb-6">
          <div className="flex items-center gap-1.5 px-4 py-2.5 border-b border-border bg-surface">
            <span className="w-2.5 h-2.5 rounded-full bg-accent-red/70" />
            <span className="w-2.5 h-2.5 rounded-full bg-accent-yellow/70" />
            <span className="w-2.5 h-2.5 rounded-full bg-accent-green/70" />
            <span className="flex-1 text-center text-2xs font-mono text-text-secondary">terminal</span>
          </div>
          <div className="px-5 py-4 flex items-center justify-between gap-4">
            <code className="font-mono text-sm text-text-secondary leading-relaxed">
              <span className="text-primary">$ </span>
              <span className="text-text-primary">{DOCKER_CMD}</span>
            </code>
            <CopyBtn text={DOCKER_CMD} />
          </div>
        </div>
      </Reveal>

      <Reveal delay={0.1}>
        <div className="space-y-3 mb-8">
          <p className="text-sm text-text-secondary leading-relaxed">
            This starts ShellHub on port 80. Open{" "}
            <code className="font-mono text-xs bg-surface px-1.5 py-0.5 rounded border border-border">http://localhost</code>{" "}
            in your browser and create your account.
          </p>
          <p className="text-sm text-text-secondary leading-relaxed">
            Then install the ShellHub agent on each device you want to manage.
          </p>
        </div>
      </Reveal>

      <Reveal delay={0.15}>
        <div className="flex items-center justify-between">
          <button
            onClick={onBack}
            className="inline-flex items-center gap-2 px-5 py-2.5 text-sm font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300 group"
          >
            <svg className="w-4 h-4 group-hover:-translate-x-0.5 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
            </svg>
            Back
          </button>

          <a
            href="https://docs.shellhub.io/getting-started"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs font-medium text-[#7B8EDB] hover:gap-2.5 transition-all group"
          >
            Full documentation
            <svg className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
            </svg>
          </a>
        </div>
      </Reveal>
    </div>
  );
}
