# Agent Guidelines for Aegisbox Portal

## Project Overview

This is a Next.js 16 application with TypeScript, using Tailwind CSS and Radix UI for UI components. The project follows shadcn/ui conventions and is configured with strict TypeScript mode.

## Build, Lint, and Test Commands

### Core Commands

```bash
pnpm dev          # Start development server (port 3000)
pnpm build        # Production build
pnpm lint         # Run ESLint on entire codebase
pnpm start        # Start production server (after build)
```

### Running ESLint

```bash
pnpm lint              # Lint all files
pnpm lint --fix        # Auto-fix linting issues
```

Note: This project uses ESLint with Next.js configuration. No test framework is currently configured.

## Code Style Guidelines

### Imports and Organization

- Use absolute imports with `@/` alias (configured in `tsconfig.json`)
- Group imports: React first, then external libraries, then internal imports
- Use `import type { Type }` for type-only imports to avoid circular dependencies
- Example:

```typescript
import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'
```

### TypeScript Conventions

- Enable strict mode (`"strict": true` in tsconfig.json)
- Use explicit return types for exported functions
- Prefer interfaces over type aliases for object shapes
- Use `Readonly<T>` for props that shouldn't be mutated
- Use `React.ComponentProps<'element'>` for component prop types

```typescript
function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: React.ComponentProps<'button'> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  })
```

### Naming Conventions

- **Components**: PascalCase for React components (`Button`, `Card`)
- **Files**: PascalCase for components, camelCase for utilities (`utils.ts`)
- **Variables/functions**: camelCase (`isLoading`, `handleSubmit`)
- **Constants**: SCREAMING_SNAKE_CASE for global constants
- **Props**: camelCase, avoid prefix like `onClick` (use `onClick` for events)
- **CSS classes**: Use Tailwind's lowercase convention (`bg-primary`, `hover:bg-accent`)

### UI Components (shadcn/ui Pattern)

- Use `class-variance-authority` (CVA) for variant handling
- Create a `variants` object with `variants` and `defaultVariants`
- Export both the component and `variants` constant
- Use Radix UI primitives for accessibility
- Use `cn()` utility for conditional class merging

```typescript
const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md...",
  {
    variants: {
      variant: { default: '...', destructive: '...' },
      size: { default: '...', sm: '...', lg: '...' },
    },
    defaultVariants: { variant: 'default', size: 'default' },
  },
)
```

### Tailwind CSS

- Use Tailwind CSS v4 with `@tailwindcss/postcss`
- Follow shadcn/ui color tokens (`bg-primary`, `text-primary-foreground`)
- Use CSS variables for theming in `app/globals.css`
- Apply `dark:` prefix for dark mode styles
- Keep utility classes in order: layout, spacing, typography, colors, effects

### Error Handling

- Use error boundaries for component-level errors
- Return early on error conditions
- Use descriptive variable names for error states (`isError`, `errorMessage`)
- Log errors appropriately with context

### Component Patterns

- Use functional components with TypeScript
- Destructure props with default values
- Use `asChild` pattern from Radix UI for component composition
- Use `data-*` attributes for component identification (`data-slot="button"`)

```typescript
function Button({ asChild = false, ...props }) {
  const Comp = asChild ? Slot : 'button'
  return <Comp data-slot="button" {...props} />
}
```

### File Structure

```
/app           - Next.js App Router pages and layouts
/components    - React components
  /ui          - shadcn/ui base components
/lib           - Utilities (utils.ts, db.ts)
/public        - Static assets
/styles        - Global styles
```

### Next.js Specifics

- Use `page.tsx` for routes, `layout.tsx` for shared layouts
- Use `Readonly<{ children: React.ReactNode }>` for layout props
- Use Next.js metadata API for SEO
- Client components: add `'use client'` at the top
- Server components: default (no directive needed)

### Additional Conventions

- No comments unless explaining complex logic
- Avoid default exports for components (named exports preferred)
- Use Zod for runtime validation with `react-hook-form` + `@hookform/resolvers`
- Use Clerk for authentication (`@clerk/nextjs`)
- Use Neon database (`@neondatabase/serverless`)

### Database Migrations

- **DO NOT run database migrations manually** - This project uses Goose for migrations
- If schema changes are needed, inform the user to run migrations via Goose
- Database schema is managed through Goose migration files in the `/migrations` directory
- Drizzle schema (`lib/schema.ts`) is for type generation only, not for migrations
