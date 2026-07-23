import { useState, useEffect } from "react";
import { cn } from "@shellhub/design-system/cn";
import { ConnectionGrid, GlowOrbs } from "@shellhub/design-system/components";
import { SiteLayout } from "@/components";
import { StepPath } from "./StepPath";
import { StepSetup } from "./StepSetup";

const steps = [{ label: "Get started" }, { label: "Setup" }];

export default function GettingStarted() {
  const [currentStep, setCurrentStep] = useState(0);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  }, [currentStep]);

  return (
    <SiteLayout className="flex flex-col">
      <main className="flex-1 flex flex-col items-center pt-24 pb-20 relative grid-bg">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="primary" />

        <div className="relative z-10 w-full max-w-4xl flex flex-col items-center">
          {/* Progress indicator */}
          <div className="flex items-center gap-3 mb-12 animate-fade-in">
            {steps.map((step, i) => (
              <div key={step.label} className="flex items-center gap-3">
                <div className="flex items-center gap-2">
                  <div
                    className={cn("w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold transition-colors duration-300", i <= currentStep ? "bg-primary text-white" : "bg-border text-text-muted")}
                  >
                    {i + 1}
                  </div>
                  <span
                    className={cn("text-xs font-medium transition-colors duration-300 hidden sm:inline", i <= currentStep ? "text-text-primary" : "text-text-muted")}
                  >
                    {step.label}
                  </span>
                </div>
                {i < steps.length - 1 && (
                  <div
                    className={cn("w-12 h-px transition-colors duration-300", i < currentStep ? "bg-primary" : "bg-border")}
                  />
                )}
              </div>
            ))}
          </div>

          {/* Step title */}
          <div
            className="text-center mb-10 animate-fade-in"
            style={{ animationDelay: "100ms" }}
          >
            <h1 className="text-[clamp(1.75rem,4vw,2.75rem)] font-bold tracking-[-0.03em] leading-tight mb-3">
              {currentStep === 0 && "Get started with ShellHub"}
              {currentStep === 1 && "Set up self-hosted"}
            </h1>
            <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed">
              {currentStep === 0 && "Choose how you want to run ShellHub."}
              {currentStep === 1 &&
                "Run ShellHub on your own infrastructure with Docker."}
            </p>
          </div>

          {/* Steps */}
          {currentStep === 0 && (
            <StepPath
              onSelectSelfHosted={() => setCurrentStep(1)}
            />
          )}
          {currentStep === 1 && (
            <StepSetup onBack={() => setCurrentStep(0)} />
          )}
        </div>
      </main>
    </SiteLayout>
  );
}
