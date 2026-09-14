import remarkFrontmatter from "remark-frontmatter";
import remarkLintHeadingIncrement from "remark-lint-heading-increment";
import remarkLintMdxJsxNoVoidChildren from "remark-lint-mdx-jsx-no-void-children";
import remarkLintMdxJsxUniqueAttributeName from "remark-lint-mdx-jsx-unique-attribute-name";
import remarkLintNoDuplicateHeadingsInSection from "remark-lint-no-duplicate-headings-in-section";
import remarkLintNoEmptyUrl from "remark-lint-no-empty-url";
import remarkMdx from "remark-mdx";
import remarkPresetLintRecommended from "remark-preset-lint-recommended";
import remarkLintDocsFrontmatter from "./scripts/remark-lint-docs-frontmatter.mjs";

export default {
  plugins: [
    remarkMdx,
    [remarkFrontmatter, ["yaml"]],
    remarkPresetLintRecommended,
    remarkLintHeadingIncrement,
    remarkLintNoDuplicateHeadingsInSection,
    remarkLintNoEmptyUrl,
    remarkLintMdxJsxUniqueAttributeName,
    remarkLintMdxJsxNoVoidChildren,
    remarkLintDocsFrontmatter,
  ],
};
