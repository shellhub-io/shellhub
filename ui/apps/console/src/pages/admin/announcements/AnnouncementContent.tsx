import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import { Markdown } from "@tiptap/markdown";
import { isAllowedUrl } from "@/utils/url";
import "@/styles/announcement-prose.css";

interface AnnouncementContentProps {
  content: string;
}

/**
 * Renders announcement content read-only, through the same editor that writes it, so the preview
 * cannot disagree with the published result.
 */
export default function AnnouncementContent({
  content,
}: AnnouncementContentProps) {
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

  if (!editor) return <div className="min-h-[120px]" />;

  return (
    <div className="announcement-prose">
      <EditorContent editor={editor} />
    </div>
  );
}
