import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { parseDocument } from "yaml";

const fields = ["layout", "title", "description"];
const optional = { schema: ["faq"] };

function report(file, reason, place) {
  const message = file.message(reason, place);
  message.source = "remark-lint";
  message.ruleId = "docs-frontmatter";
}

export default function remarkLintDocsFrontmatter() {
  return (tree, file) => {
    const nodes = tree.children.filter((node) => node.type === "yaml");

    if (nodes.length !== 1) {
      report(file, "Expected exactly one YAML frontmatter block", tree);
      return;
    }

    const node = nodes[0];
    const document = parseDocument(node.value, { uniqueKeys: true });

    for (const error of document.errors) {
      report(file, error.message, node);
    }

    if (document.errors.length > 0) return;

    const frontmatter = document.toJS();

    if (frontmatter === null || Array.isArray(frontmatter) || typeof frontmatter !== "object") {
      report(file, "Expected frontmatter to be a mapping", node);
      return;
    }

    for (const field of fields) {
      if (typeof frontmatter[field] !== "string" || frontmatter[field].trim() === "") {
        report(file, `Expected a non-empty ${field} string`, node);
      }
    }

    for (const [field, value] of Object.entries(frontmatter)) {
      if (Object.hasOwn(optional, field)) {
        if (!optional[field].includes(value)) {
          report(file, `Expected ${field} to be one of ${optional[field].join(", ")}`, node);
        }
      } else if (!fields.includes(field)) {
        report(file, `Unexpected frontmatter field ${field}`, node);
      }
    }

    if (
      typeof frontmatter.layout === "string" &&
      !existsSync(resolve(dirname(file.path), frontmatter.layout))
    ) {
      report(file, `Layout does not exist: ${frontmatter.layout}`, node);
    }
  };
}
