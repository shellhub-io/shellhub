import { useId } from "react";
import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import { Markdown } from "@tiptap/markdown";
import { MegaphoneIcon, XMarkIcon } from "@heroicons/react/24/outline";
import {
  Button,
  IconBadge,
  IconButton,
} from "@shellhub/design-system/primitives";
import BaseDialog from "@/components/common/BaseDialog";
import { formatDateShort } from "@/utils/date";
import { isAllowedUrl } from "@/utils/url";
import type { Announcement } from "@/client";
import "./AnnouncementModal.css";

interface AnnouncementContentProps {
  content: string;
}

function AnnouncementContent({ content }: AnnouncementContentProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Link.configure({
        openOnClick: true,
        validate: (url) => isAllowedUrl(url),
        HTMLAttributes: {
          rel: "noopener noreferrer",
          target: "_blank",
        },
      }),
      Image.configure({ allowBase64: false }),
      Markdown,
    ],
    content,
    contentType: "markdown",
    editable: false,
  });

  if (!editor) return <div className="min-h-[80px]" />;

  return (
    <div className="announcement-modal-content">
      <EditorContent editor={editor} />
    </div>
  );
}

interface AnnouncementModalProps {
  open: boolean;
  onClose: () => void;
  announcement: Announcement;
}

export default function AnnouncementModal({
  open,
  onClose,
  announcement,
}: AnnouncementModalProps) {
  const titleId = useId();

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="md"
      aria-labelledby={titleId}
    >
      <div className="flex items-start justify-between gap-4 p-6 border-b border-border">
        <div className="flex items-center gap-3">
          <IconBadge size="md">
            <MegaphoneIcon className="w-5 h-5 text-primary" strokeWidth={1.5} />
          </IconBadge>
          <div>
            <h2
              id={titleId}
              className="text-base font-semibold text-text-primary leading-snug"
            >
              {announcement.title}
            </h2>
            <p className="text-xs text-text-muted font-mono mt-0.5">
              {formatDateShort(announcement.date)}
            </p>
          </div>
        </div>
        <IconButton
          size="sm"
          onClick={onClose}
          aria-label="Close announcement"
          className="-mt-0.5 -mr-1"
        >
          <XMarkIcon className="w-4 h-4" />
        </IconButton>
      </div>

      <div className="p-6 overflow-y-auto max-h-[60vh]">
        <AnnouncementContent
          key={announcement.uuid}
          content={announcement.content}
        />
      </div>

      <div className="flex justify-end gap-2 p-5 border-t border-border">
        <Button variant="ghost" onClick={onClose}>
          Got it
        </Button>
      </div>
    </BaseDialog>
  );
}
