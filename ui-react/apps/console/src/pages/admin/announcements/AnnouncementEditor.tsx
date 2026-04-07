import { useCallback, useRef, type MutableRefObject } from "react";
import { useEditor, EditorContent, type Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import { Markdown } from "@tiptap/markdown";
import { isAllowedUrl } from "@/utils/url";
import "./AnnouncementEditor.css";

interface AnnouncementEditorProps {
  content: string;
  onChange: (markdown: string) => void;
  editorRef?: MutableRefObject<Editor | null>;
}

function ToolbarButton({
  onClick,
  active,
  disabled,
  title,
  children,
}: {
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={title}
      aria-label={title}
      aria-pressed={active}
      className={`p-1.5 rounded text-xs transition-colors ${
        active
          ? "bg-primary/15 text-primary"
          : "text-text-muted hover:text-text-primary hover:bg-surface"
      } disabled:opacity-30 disabled:cursor-not-allowed`}
    >
      {children}
    </button>
  );
}

function ToolbarSeparator() {
  return <div className="w-px h-5 bg-border mx-0.5" />;
}

function EditorToolbar({ editor }: { editor: Editor }) {
  const toolbarRef = useRef<HTMLDivElement>(null);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    const buttons = toolbarRef.current?.querySelectorAll<HTMLButtonElement>(
      "button:not([disabled])",
    );
    if (!buttons?.length) return;
    const currentIndex = Array.from(buttons).indexOf(
      document.activeElement as HTMLButtonElement,
    );
    if (currentIndex === -1) return;

    let nextIndex: number | null = null;
    if (e.key === "ArrowRight") nextIndex = (currentIndex + 1) % buttons.length;
    if (e.key === "ArrowLeft")
      nextIndex = (currentIndex - 1 + buttons.length) % buttons.length;

    if (nextIndex !== null) {
      e.preventDefault();
      buttons[nextIndex].focus();
    }
  };

  const setLink = useCallback(() => {
    const previous = editor.getAttributes("link").href as string | undefined;
    const url = window.prompt("URL", previous);
    if (url === null) return;
    if (url === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      return;
    }
    if (!isAllowedUrl(url)) return;
    editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
  }, [editor]);

  const addImage = useCallback(() => {
    const url = window.prompt("Image URL");
    if (url && isAllowedUrl(url)) {
      editor.chain().focus().setImage({ src: url }).run();
    }
  }, [editor]);

  return (
    <div
      ref={toolbarRef}
      className="flex items-center gap-0.5 px-2 py-1.5 border-b border-border bg-surface/30 flex-wrap"
      role="toolbar"
      aria-label="Text formatting"
      onKeyDown={handleKeyDown}
    >
      {/* Text formatting */}
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleBold().run()}
        active={editor.isActive("bold")}
        title="Bold"
      >
        <span className="font-bold">B</span>
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleItalic().run()}
        active={editor.isActive("italic")}
        title="Italic"
      >
        <span className="italic">I</span>
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleStrike().run()}
        active={editor.isActive("strike")}
        title="Strikethrough"
      >
        <span className="line-through">S</span>
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleCode().run()}
        active={editor.isActive("code")}
        title="Inline code"
      >
        <span className="font-mono">&lt;/&gt;</span>
      </ToolbarButton>

      <ToolbarSeparator />

      {/* Headings */}
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
        active={editor.isActive("heading", { level: 1 })}
        title="Heading 1"
      >
        H1
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
        active={editor.isActive("heading", { level: 2 })}
        title="Heading 2"
      >
        H2
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
        active={editor.isActive("heading", { level: 3 })}
        title="Heading 3"
      >
        H3
      </ToolbarButton>

      <ToolbarSeparator />

      {/* Lists */}
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleBulletList().run()}
        active={editor.isActive("bulletList")}
        title="Bullet list"
      >
        &bull; List
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleOrderedList().run()}
        active={editor.isActive("orderedList")}
        title="Numbered list"
      >
        1. List
      </ToolbarButton>

      <ToolbarSeparator />

      {/* Block elements */}
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleBlockquote().run()}
        active={editor.isActive("blockquote")}
        title="Blockquote"
      >
        &ldquo; Quote
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().toggleCodeBlock().run()}
        active={editor.isActive("codeBlock")}
        title="Code block"
      >
        {"{ }"}
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().setHorizontalRule().run()}
        title="Horizontal rule"
      >
        &mdash;
      </ToolbarButton>

      <ToolbarSeparator />

      {/* Links and images */}
      <ToolbarButton
        onClick={setLink}
        active={editor.isActive("link")}
        title="Link"
      >
        Link
      </ToolbarButton>
      <ToolbarButton onClick={addImage} title="Image">
        Image
      </ToolbarButton>

      <ToolbarSeparator />

      {/* Undo/Redo */}
      <ToolbarButton
        onClick={() => editor.chain().focus().undo().run()}
        disabled={!editor.can().undo()}
        title="Undo"
      >
        ↩
      </ToolbarButton>
      <ToolbarButton
        onClick={() => editor.chain().focus().redo().run()}
        disabled={!editor.can().redo()}
        title="Redo"
      >
        ↪
      </ToolbarButton>
    </div>
  );
}

export default function AnnouncementEditor({
  content,
  onChange,
  editorRef,
}: AnnouncementEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Link.configure({
        openOnClick: false,
        defaultProtocol: "https",
        validate: (url) => isAllowedUrl(url),
      }),
      Image.configure({ allowBase64: false }),
      Markdown,
    ],
    content,
    contentType: "markdown",
    onUpdate: ({ editor: e }) => {
      onChange(e.getMarkdown());
    },
    onCreate: ({ editor: e }) => {
      if (editorRef) editorRef.current = e;
    },
    onDestroy: () => {
      if (editorRef) editorRef.current = null;
    },
    editorProps: {
      attributes: {
        id: "announcement-content-editor",
        role: "textbox",
        "aria-multiline": "true",
        "aria-label": "Announcement content",
      },
    },
  });

  if (!editor) return null;

  return (
    <div className="announcement-editor border border-border rounded-lg overflow-hidden bg-card">
      <EditorToolbar editor={editor} />
      <EditorContent editor={editor} />
    </div>
  );
}
