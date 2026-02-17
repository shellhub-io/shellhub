import { defineCollection, z } from "astro:content";
import { glob } from "astro/loaders";

const blog = defineCollection({
  loader: glob({ pattern: "**/*.mdx", base: "./src/content/blog" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    author: z.string(),
    authorRole: z.string().optional(),
    authorAvatar: z.string().optional(),
    date: z.coerce.date(),
    categories: z.array(z.string()).default([]),
    tags: z.array(z.string()).default([]),
    image: z.string().optional(),
    ogTagline: z.string().optional(),
    ogIcons: z.array(z.string()).default([]),
    ogLayout: z.enum(["default", "integration"]).default("default"),
    ogPartnerIcon: z.string().optional(),
    draft: z.boolean().default(false),
  }),
});

export const collections = { blog };
