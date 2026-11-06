import type { Nodes, Root } from "mdast";
import { fromMarkdown } from "mdast-util-from-markdown";
import { gfmFromMarkdown } from "mdast-util-gfm";
import { gfm } from "micromark-extension-gfm";

/** Parses GitHub-flavoured Markdown, so a test can assert on the tree instead of the text. */
export const parse = (markdown: string): Root =>
  fromMarkdown(markdown, { extensions: [gfm()], mdastExtensions: [gfmFromMarkdown()] });

/** Every node of `type` under `tree`, the tree itself included, in document order. */
export const nodesOf = <T extends Nodes>(tree: Nodes, type: T["type"]): T[] => [
  ...(tree.type === type ? [tree as T] : []),
  ...("children" in tree ? tree.children.flatMap((child) => nodesOf<T>(child, type)) : []),
];
