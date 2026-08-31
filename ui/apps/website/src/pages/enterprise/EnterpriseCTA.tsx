import { CTABanner } from "@/components";

/**
 * Closing call to action of the enterprise page. The primary action is a mailto rather than a
 * route: there is no contact form, and sales is reached by mail.
 */
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
