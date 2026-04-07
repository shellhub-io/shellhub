import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import { Markdown } from "@tiptap/markdown";
import { isAllowedUrl } from "@/utils/url";
import "./AnnouncementEditor.css";

interface AnnouncementContentProps {
  content: string;
}

export default function AnnouncementContent({
  content,
}: AnnouncementContentProps) {
  // content is only used on initial mount. Parent should use a key prop
  // (e.g. key={announcement.uuid}) to force re-mount on content changes.
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
    <div className="announcement-editor">
      <EditorContent editor={editor} />
    </div>
  );
}
