import { useEffect, useState, useCallback } from "react";
import { Navbar } from "./Navbar";
import { Hero } from "./Hero";
import { TrustedBy } from "./TrustedBy";
import { QuickStart } from "./QuickStart";
import { SupportedPlatforms } from "./SupportedPlatforms";
import { HowItWorks } from "./HowItWorks";
import { Architecture } from "./Architecture";
import { FeatureGrid } from "./FeatureGrid";
import { OpenSource } from "./OpenSource";
import { CTA } from "./CTA";
import { Footer } from "@shellhub/design-system/components";

export default function Landing() {
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);

  const handleScroll = useCallback(() => {
    setNavSolid(window.scrollY > 50);
  }, []);

  useEffect(() => {
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar
        navSolid={navSolid}
        mobileMenu={mobileMenu}
        setMobileMenu={setMobileMenu}
      />
      <Hero />
      <TrustedBy />
      <QuickStart />
      <SupportedPlatforms />
      <HowItWorks />
      <Architecture />
      <FeatureGrid />
      <OpenSource />
      <CTA />
      <Footer />
    </div>
  );
}
