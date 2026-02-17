import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { HeroEnterprise } from "./HeroEnterprise";
import { AdminPanel } from "./AdminPanel";
import { SecurityFeatures } from "./SecurityFeatures";
import { DeploymentOptions } from "./DeploymentOptions";
import { SupportSection } from "./SupportSection";
import { EnterpriseCTA } from "./EnterpriseCTA";

export default function Enterprise() {
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
      <HeroEnterprise />
      <AdminPanel />
      <SecurityFeatures />
      <DeploymentOptions />
      <SupportSection />
      <EnterpriseCTA />
      <Footer />
    </div>
  );
}
