import { CTABanner } from "@/components/marketing";

export function EnterpriseCTA() {
  return (
    <CTABanner
      eyebrow="Ready to get started?"
      title="Talk to our team"
      subtitle="Get a demo, discuss your requirements, and find the right plan for your organization. Our team typically responds within one business day."
      primaryAction={{
        label: "Contact Sales",
        href: "mailto:sales@shellhub.io",
      }}
      secondaryAction={{ label: "View Pricing", to: "/pricing" }}
    />
  );
}
