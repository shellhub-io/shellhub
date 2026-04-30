import { useState } from "react";
import { getConfig } from "@/env";
import { useLatestAnnouncement } from "@/hooks/useLatestAnnouncement";
import AnnouncementModal from "./AnnouncementModal";
import type { Announcement } from "@/client";

const STORAGE_KEY = "announcement";

function computeHash(announcement: Announcement): string {
  const json = JSON.stringify(announcement);
  return btoa(
    Array.from(new TextEncoder().encode(json), (b) =>
      String.fromCharCode(b),
    ).join(""),
  );
}

function getStoredHash(): string {
  return localStorage.getItem(STORAGE_KEY) ?? "";
}

function markSeen(announcement: Announcement): void {
  localStorage.setItem(STORAGE_KEY, computeHash(announcement));
}

export default function AnnouncementModalTrigger() {
  if (!getConfig().announcements) return null;

  return <AnnouncementModalInner />;
}

function AnnouncementModalInner() {
  const { announcement } = useLatestAnnouncement();
  const [dismissed, setDismissed] = useState(false);

  const show =
    !!announcement &&
    !dismissed &&
    computeHash(announcement) !== getStoredHash();

  const handleClose = () => {
    if (announcement) markSeen(announcement);
    setDismissed(true);
  };

  if (!show || !announcement) return null;

  return (
    <AnnouncementModal open={show} onClose={handleClose} announcement={announcement} />
  );
}
