import { useState, useEffect } from "react";
import { Navbar } from "@/pages/landing/Navbar";
import { Footer } from "@shellhub/design-system/components";

export function SiteLayout({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);

  const handleScroll = () => {
    setNavSolid(window.scrollY > 50);
  };

  useEffect(() => {
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <div
      className={`min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden${className ? " " + className : ""}`}
    >
      <Navbar
        navSolid={navSolid}
        mobileMenu={mobileMenu}
        setMobileMenu={setMobileMenu}
      />
      {children}
      <Footer />
    </div>
  );
}
