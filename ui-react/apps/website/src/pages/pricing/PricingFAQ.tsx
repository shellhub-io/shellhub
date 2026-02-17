import { useState } from "react";
import { Reveal } from "../landing/components";

const faqs = [
  {
    q: "Can I switch plans later?",
    a: "Yes. You can upgrade from Community to Cloud or Enterprise at any time. Your devices and configuration carry over seamlessly.",
  },
  {
    q: "What counts as a device?",
    a: "A device is any endpoint running the ShellHub agent that has connected to your namespace. Devices that are removed from your namespace no longer count toward your limit.",
  },
  {
    q: "Is there a free trial for Cloud?",
    a: "Cloud is free for up to 3 devices with no time limit. You only start paying when you add more than 3 devices to your namespace.",
  },
  {
    q: "Can I self-host the Enterprise edition?",
    a: "Yes. Enterprise supports both managed cloud and on-premises deployments. You can run it on your own Kubernetes cluster or with Docker Compose.",
  },
  {
    q: "What kind of support does Enterprise include?",
    a: "Enterprise includes a dedicated account manager, priority ticket queue with SLA guarantees, private communication channels, and onboarding assistance.",
  },
  {
    q: "Is the Community edition really free?",
    a: "Yes. ShellHub Community is open source under the Apache 2.0 license. You can self-host it at no cost for up to 3 devices.",
  },
];

function FAQItem({ q, a, index }: { q: string; a: string; index: number }) {
  const [open, setOpen] = useState(false);

  return (
    <Reveal delay={index * 0.04}>
      <button
        onClick={() => setOpen(!open)}
        className="w-full text-left py-5 border-b border-border group"
      >
        <div className="flex items-center justify-between gap-4">
          <h4 className="text-sm font-semibold group-hover:text-primary transition-colors">{q}</h4>
          <svg
            className={`w-4 h-4 text-text-muted shrink-0 transition-transform duration-300 ${open ? "rotate-45" : ""}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
        </div>
        <div
          className={`overflow-hidden transition-all duration-300 ${
            open ? "max-h-40 opacity-100 mt-3" : "max-h-0 opacity-0"
          }`}
        >
          <p className="text-sm text-text-secondary leading-relaxed pr-8">{a}</p>
        </div>
      </button>
    </Reveal>
  );
}

export function PricingFAQ() {
  return (
    <section className="py-24">
      <div className="max-w-3xl mx-auto px-8">
        <Reveal className="text-center mb-14">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
            FAQ
          </p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight">
            Frequently asked questions
          </h2>
        </Reveal>

        <div>
          {faqs.map((faq, i) => (
            <FAQItem key={i} q={faq.q} a={faq.a} index={i} />
          ))}
        </div>
      </div>
    </section>
  );
}
