# Deploy Plan — ShellHub Website & Docs (Cloudflare Pages)

## Contexto

O site (website + blog) e docs são apps estáticos dentro do monorepo ShellHub. Hoje tudo roda via Docker Compose junto com o app principal. O objetivo é deployar o site estático separadamente via Cloudflare Pages, independente do ciclo de release do produto.

## Arquitetura

```
shellhub/ (monorepo)
│
├── ui-react/apps/website   → Cloudflare Pages (shellhub.io)
├── ui-react/apps/docs      → Cloudflare Pages (docs.shellhub.io)
│
├── api/                    ─┐
├── ssh/                     │→ Docker Compose (app.shellhub.io)
├── gateway/                 │   Deploy separado via release tags
├── ui/                     ─┘
```

Cada um tem seu próprio ciclo de deploy. Publicar um blog post não rebuilda o app. Lançar versão nova do app não rebuilda o site.

## Cloudflare Pages — Configuração

### Projeto 1: shellhub-website

| Campo | Valor |
|---|---|
| Nome do projeto | `shellhub-website` |
| Repositório | `shellhub-io/shellhub` |
| Branch de produção | `main` |
| Root directory | `ui-react` |
| Build command | `npm run build --workspace=apps/website` |
| Output directory | `apps/website/dist` |
| Custom domain | `shellhub.io` |
| Node version (env var) | `NODE_VERSION=22` |

### Projeto 2: shellhub-docs

| Campo | Valor |
|---|---|
| Nome do projeto | `shellhub-docs` |
| Repositório | `shellhub-io/shellhub` |
| Branch de produção | `main` |
| Root directory | `ui-react` |
| Build command | `npm run build --workspace=apps/docs` |
| Output directory | `apps/docs/dist` |
| Custom domain | `docs.shellhub.io` |
| Node version (env var) | `NODE_VERSION=22` |

## Fluxo de Deploy

### Blog post / conteúdo do site

```
Writer cria post.mdx no repo
        ↓
Abre PR → Cloudflare gera preview deploy automático
          (URL temporária tipo abc123.shellhub-website.pages.dev)
        ↓
Review do conteúdo no preview → merge na main
        ↓
Cloudflare detecta push na main
        ↓
Builda apps/website → deploy em ~30-60s
        ↓
Live em shellhub.io/blog/slug
```

### Docs

Mesmo fluxo. Mudanças em `apps/docs/` trigeram rebuild do projeto shellhub-docs.

### App principal (API, SSH, Gateway, UI)

Fluxo independente — continua via Docker Compose / release tags como hoje. Cloudflare Pages não é afetado.

## Blog — Estrutura de Conteúdo

O blog usa arquivos MDX no repositório (mesma abordagem do Supabase):

```
ui-react/apps/website/
├── src/
│   ├── pages/
│   │   └── blog/
│   │       ├── index.astro          # Listagem de posts
│   │       └── [slug].astro         # Post individual
│   ├── layouts/
│   │   └── BlogLayout.astro         # Layout do post
│   └── components/
│       └── blog/                    # Componentes do blog
├── content/
│   └── blog/                        # Posts MDX
│       ├── 2026-01-15-introducing-shellhub-v1.mdx
│       ├── 2026-02-01-remote-access-best-practices.mdx
│       └── ...
└── astro.config.ts                  # Content collections config
```

### Frontmatter dos posts

```yaml
---
title: "Post Title"
description: "Brief description for SEO and cards"
author: "gustavo"
date: 2026-02-13
categories:
  - product
tags:
  - ssh
  - security
image: "./cover.png"
draft: false
---
```

### Astro Content Collections

O Astro tem suporte nativo a content collections com validação de schema via Zod:

```ts
// src/content.config.ts
import { defineCollection, z } from "astro:content";

const blog = defineCollection({
  type: "content",
  schema: z.object({
    title: z.string(),
    description: z.string(),
    author: z.string(),
    date: z.date(),
    categories: z.array(z.string()).optional(),
    tags: z.array(z.string()).optional(),
    image: z.string().optional(),
    draft: z.boolean().default(false),
  }),
});

export const collections = { blog };
```

## Cloudflare Pages — Recursos Gratuitos

- Builds ilimitados
- 500 deploys/mês
- Preview deploys automáticos por PR
- CDN global com edge caching
- Custom domains com SSL automático
- Bandwidth ilimitado
- Support a monorepos

## Considerações

### Build no monorepo

Cloudflare Pages faz `npm install` a partir do root directory configurado (`ui-react`). Como usamos npm workspaces, o install resolve dependências de todos os apps. O build command com `--workspace=apps/website` builda apenas o app desejado.

### Variáveis de ambiente

Configurar no dashboard do Cloudflare Pages:
- `NODE_VERSION=22`
- Qualquer `VITE_SHELLHUB_*` necessária (ex: `VITE_SHELLHUB_CLOUD=true`)

### Preview deploys

Cada PR gera automaticamente um deploy de preview com URL única. Útil pra revisar blog posts e mudanças no site antes de mergear.

### Rollback

Cloudflare mantém histórico de deploys. Rollback instantâneo pra qualquer versão anterior via dashboard.

## TODO

- [ ] Configurar projeto shellhub-website no Cloudflare Pages
- [ ] Configurar projeto shellhub-docs no Cloudflare Pages
- [ ] Configurar custom domains e DNS
- [ ] Implementar blog no website app (Astro content collections)
- [ ] Criar layout do blog (listagem + post individual)
- [ ] Migrar conteúdo existente do blog atual (se houver)
- [ ] Testar preview deploys com PR de teste
- [ ] Configurar redirects (se necessário, via `_redirects` file)
