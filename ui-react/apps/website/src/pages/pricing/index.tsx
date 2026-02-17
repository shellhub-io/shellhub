import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { HeroPricing } from "./HeroPricing";
import { TierCards } from "./TierCards";
import { ComparisonTable } from "./ComparisonTable";
import { PricingFAQ } from "./PricingFAQ";

export default function Pricing() {
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);

  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);

  useEffect(() => {
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />
      <HeroPricing />
      <TierCards />
      <ComparisonTable />
      <PricingFAQ />
      <Footer />
    </div>
  );
}
