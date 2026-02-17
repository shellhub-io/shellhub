import { useEffect, useState, useCallback } from "react";
import { Footer, ConnectionGrid } from "@shellhub/design-system/components";
import { Navbar } from "../landing/Navbar";
import { StepPath } from "./StepPath";
import { StepSetup } from "./StepSetup";
import { StepSignup } from "./StepSignup";

type Path = "cloud" | "selfhosted";

const steps = [
  { label: "Get started" },
  { label: "Setup" },
];

export default function GettingStarted() {
  const [currentStep, setCurrentStep] = useState(0);
  const [path, setPath] = useState<Path | null>(null);
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);

  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);

  useEffect(() => {
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  }, [currentStep]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden flex flex-col">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />

      <main className="flex-1 flex flex-col items-center pt-24 pb-20 relative grid-bg">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/10 via-transparent to-transparent pointer-events-none" />

        <div className="relative z-10 w-full max-w-4xl flex flex-col items-center">
          {/* Progress indicator */}
          <div className="flex items-center gap-3 mb-12 animate-fade-in">
            {steps.map((step, i) => (
              <div key={step.label} className="flex items-center gap-3">
                <div className="flex items-center gap-2">
                  <div className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold transition-colors duration-300 ${
                    i <= currentStep
                      ? "bg-primary text-white"
                      : "bg-border text-text-muted"
                  }`}>
                    {i + 1}
                  </div>
                  <span className={`text-xs font-medium transition-colors duration-300 hidden sm:inline ${
                    i <= currentStep ? "text-text-primary" : "text-text-muted"
                  }`}>
                    {step.label}
                  </span>
                </div>
                {i < steps.length - 1 && (
                  <div className={`w-12 h-px transition-colors duration-300 ${
                    i < currentStep ? "bg-primary" : "bg-border"
                  }`} />
                )}
              </div>
            ))}
          </div>

          {/* Step title */}
          <div className="text-center mb-10 animate-fade-in" style={{ animationDelay: "100ms" }}>
            <h1 className="text-[clamp(1.75rem,4vw,2.75rem)] font-bold tracking-[-0.03em] leading-tight mb-3">
              {currentStep === 0 && "Get started with ShellHub"}
              {currentStep === 1 && path === "selfhosted" && "Set up self-hosted"}
              {currentStep === 1 && path === "cloud" && "Create your account"}
            </h1>
            <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed">
              {currentStep === 0 && "Choose how you want to run ShellHub."}
              {currentStep === 1 && path === "selfhosted" && "Run ShellHub on your own infrastructure with Docker."}
              {currentStep === 1 && path === "cloud" && "Sign up to start using ShellHub Cloud."}
            </p>
          </div>

          {/* Steps */}
          {currentStep === 0 && (
            <StepPath
              onSelectCloud={() => { setPath("cloud"); setCurrentStep(1); }}
              onSelectSelfHosted={() => { setPath("selfhosted"); setCurrentStep(1); }}
            />
          )}
          {currentStep === 1 && path === "selfhosted" && (
            <StepSetup onBack={() => setCurrentStep(0)} />
          )}
          {currentStep === 1 && path === "cloud" && (
            <StepSignup onBack={() => setCurrentStep(0)} />
          )}
        </div>
      </main>

      <Footer />
    </div>
  );
}
