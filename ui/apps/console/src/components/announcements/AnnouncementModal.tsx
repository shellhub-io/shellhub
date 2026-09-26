import { useId } from "react";
import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import { Markdown } from "@tiptap/markdown";
import { MegaphoneIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import BaseDialog from "@/components/common/BaseDialog";
import { formatDateShort } from "@/utils/date";
import { isAllowedUrl } from "@/utils/url";
import type { Announcement } from "@/client";
import "@/styles/announcement-prose.css";
import DialogHeader from "@/components/common/DialogHeader";

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
    <div className="announcement-prose">
      <EditorContent editor={editor} />
    </div>
  );
}

interface AnnouncementModalProps {
  open: boolean;
  onClose: () => void;
  announcement: Announcement;
}

/**
 * Shows one announcement in full. The body is rendered from Markdown, so the source is trusted
 * content written by an instance admin rather than by a user.
 */
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
      aria-describedby={`${titleId}-description`}
    >
      <DialogHeader
        icon={<MegaphoneIcon />}
        title={announcement.title}
        description={
          <span className="font-mono">
            {formatDateShort(announcement.date)}
          </span>
        }
        titleId={titleId}
        descriptionId={`${titleId}-description`}
        onClose={onClose}
      />

      <div className="p-6 overflow-y-auto max-h-[60vh] border-t border-border">
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
