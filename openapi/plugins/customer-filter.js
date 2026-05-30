// Redocly decorator that keeps only the customer-facing surface of the API.
//
// An operation is kept when it is usable by a customer integration, which means
// all of:
//   1. its path is not under /admin or /internal (those surfaces are never
//      customer facing), and
//   2. it accepts the `api-key` security scheme (api-key auth is namespace
//      related and not tied to a user), and
//   3. it is not explicitly flagged with `x-internal: true`.
//
// Everything else (admin, internal, login/auth, account, billing, MFA, user and
// the few api-key accepting but account level routes) is dropped. Path items
// left without any operation are removed too.
//
// Operations are removed at the operation level, because deleting a whole path
// item from its parent does not stick for paths composed through a
// `paths: $ref` import (cloud and enterprise specs). Empty path items are then
// pruned in the Root visitor, which runs on the fully assembled document.
//
// The `x-internal` flag is an explicit override for operations that accept an
// api-key but are still not part of the customer integration surface (for
// example create/list namespace, leave namespace, list api keys, member
// invitations).

// Home page (info.description) for the customer docs. The full spec keeps its
// own internal-facing description; this one is injected only into the filtered
// build.
const CUSTOMER_DESCRIPTION = `Programmatic access to your namespace: devices, sessions, SSH public keys,
tags, firewall rules, and more.

## Base URL

Endpoints are served under \`/api\` on your ShellHub server. On ShellHub Cloud
the base URL is \`https://cloud.shellhub.io/api\`.

## Authentication

Send your API key in the \`X-API-KEY\` header:

\`\`\`
curl https://cloud.shellhub.io/api/devices -H "X-API-KEY: <your-key>"
\`\`\`

An API key belongs to a single namespace and is not tied to a user. Create one
in the console under **Namespace → API Keys**. Every endpoint operates within
the key's namespace, so the namespace scope is implicit throughout this
reference.

## Pagination

List endpoints accept \`page\` and \`per_page\` query parameters and return the
total item count in the \`X-Total-Count\` response header.

## Errors

Errors use standard HTTP status codes. \`401\` means the API key is missing or
invalid; \`403\` means the key's role does not allow the operation.`;

const HTTP_METHODS = [
  'get',
  'put',
  'post',
  'delete',
  'options',
  'head',
  'patch',
  'trace',
];

const INTERNAL_PREFIXES = ['/admin', '/internal'];

// Edition and audience markers used as tags across the specs. They are not
// resources, so they are stripped from the customer docs to leave a clean
// resource-based grouping (devices, sessions, rules, ...).
const NON_RESOURCE_TAGS = new Set([
  'community',
  'cloud',
  'enterprise',
  'internal',
  'external',
]);

function operationAcceptsApiKey(operation) {
  const security = operation.security;

  return (
    Array.isArray(security) &&
    security.some(
      (requirement) =>
        requirement &&
        Object.prototype.hasOwnProperty.call(requirement, 'api-key'),
    )
  );
}

function isCustomerOperation(operation) {
  return operation['x-internal'] !== true && operationAcceptsApiKey(operation);
}

function hasAnyOperation(pathItem) {
  return HTTP_METHODS.some((method) => pathItem[method] !== undefined);
}

function DropNonCustomer() {
  return {
    PathItem: {
      enter(pathItem, ctx) {
        const path = ctx.key;
        const isInternalPath =
          typeof path === 'string' &&
          INTERNAL_PREFIXES.some((prefix) => path.startsWith(prefix));

        for (const method of HTTP_METHODS) {
          const operation = pathItem[method];

          if (operation === undefined) {
            continue;
          }

          if (isInternalPath || !isCustomerOperation(operation)) {
            delete pathItem[method];

            continue;
          }

          // Kept operation: clean it up for the customer docs.
          if (Array.isArray(operation.tags)) {
            operation.tags = operation.tags.filter(
              (tag) => !NON_RESOURCE_TAGS.has(tag),
            );
          }

          // The customer docs are api-key only, so drop the jwt alternative
          // from the security requirements (it stays in the full spec).
          if (Array.isArray(operation.security)) {
            operation.security = operation.security.filter(
              (requirement) =>
                requirement &&
                Object.prototype.hasOwnProperty.call(requirement, 'api-key'),
            );
          }
        }
      },
    },
    Root: {
      leave(root) {
        if (!root.paths) {
          return;
        }

        for (const path of Object.keys(root.paths)) {
          if (!hasAnyOperation(root.paths[path])) {
            delete root.paths[path];
          }
        }

        // Drop root tags that no surviving operation references, so the
        // rendered docs do not show empty tag groups.
        if (Array.isArray(root.tags)) {
          const usedTags = new Set();

          for (const pathItem of Object.values(root.paths)) {
            for (const method of HTTP_METHODS) {
              const operation = pathItem[method];

              if (operation && Array.isArray(operation.tags)) {
                operation.tags.forEach((tag) => usedTags.add(tag));
              }
            }
          }

          const seenTags = new Set();
          root.tags = root.tags.filter((tag) => {
            if (!usedTags.has(tag.name) || seenTags.has(tag.name)) {
              return false;
            }

            seenTags.add(tag.name);

            return true;
          });
        }

        // The customer docs authenticate only with api-key, so drop the jwt
        // security scheme. The full spec keeps it.
        if (root.components && root.components.securitySchemes) {
          delete root.components.securitySchemes.jwt;
        }

        // Replace the home page with the customer-facing description. The full
        // spec keeps its internal-facing one.
        if (root.info) {
          root.info.description = CUSTOMER_DESCRIPTION;
        }
      },
    },
  };
}

function customerFilterPlugin() {
  return {
    id: 'customer-filter',
    decorators: {
      oas3: {
        'drop-non-customer': DropNonCustomer,
      },
    },
  };
}

module.exports = customerFilterPlugin;
