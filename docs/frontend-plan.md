# BD Govt Job Circular — Frontend Implementation Plan

> **Stack:** Next.js 14+ · TypeScript · Tailwind CSS · shadcn/ui · TanStack Query  
> **Backend:** Go API (`http://localhost:8080/api/v1`)  
> **Status:** Planning — ready to implement

---

## Table of Contents

1. [Tech Stack](#1-tech-stack)
2. [Pages & Routes](#2-pages--routes)
3. [Component Architecture](#3-component-architecture)
4. [Auth Flow (Detailed)](#4-auth-flow-detailed)
5. [Data Fetching Strategy](#5-data-fetching-strategy)
6. [Feature Deep Dives](#6-feature-deep-dives)
7. [Admin Dashboard](#7-admin-dashboard)
8. [File / Asset Handling (MinIO)](#8-file--asset-handling-minio)
9. [SEO & Performance](#9-seo--performance)
10. [Internationalization (Bangla)](#10-internationalization-bangla)
11. [Environment Variables](#11-environment-variables)
12. [Project Structure](#12-project-structure)
13. [Deployment](#13-deployment)
14. [Implementation Phases](#14-implementation-phases)

---

## 1. Tech Stack

| Layer | Choice | Why |
|---|---|---|
| Framework | **Next.js 14+** (App Router) | SSR/SSG for SEO, server components for data fetching, API routes for proxy |
| Language | **TypeScript** | Type safety across API contracts |
| Styling | **Tailwind CSS** + **shadcn/ui** | Fast UI building, accessible components, Bangla-friendly |
| Data Fetching | **TanStack Query v5** | Caching, pagination, optimistic updates for bookmarks |
| Auth State | **Zustand** or React Context | Lightweight JWT token management |
| Form Handling | **React Hook Form** + **Zod** | Validation for register, login, profile, alerts |
| HTTP Client | **fetch + custom API client** | Lightweight, built-in to Next.js |
| Icons | **Lucide React** | Clean, consistent icon set |
| Analytics | **Vercel Analytics** or Plausible | Page views, circular views, conversion tracking |
| Deployment | **Vercel** or Docker on same server | Your backend runs on DO, frontend can too |

---

## 2. Pages & Routes

### Public Pages (no auth required)

| Route | Page | Backend API | Description |
|---|---|---|---|
| `/` | Home | `GET /circulars` (featured) | Hero, featured circulars, search bar, category pills |
| `/circulars` | Circular Listing | `GET /circulars` | Full list with filters, pagination, search |
| `/circulars/[id]` | Circular Detail | `GET /circulars/:id` | Full detail, apply button, bookmark button |
| `/categories` | Categories | `GET /categories` | Grid of 10 categories with icons |
| `/categories/[slug]` | Category Filter | `GET /circulars?category=slug` | Circulars filtered by category |
| `/organizations` | Organizations | `GET /organizations` | Searchable list of all organizations |
| `/login` | Login | `POST /auth/login` | Email + password form |
| `/register` | Register | `POST /auth/register` | Registration form with optional fields |
| `/verify-email` | Email Verification | `GET /auth/verify-email?token=` | Token from email link → verify → redirect |
| `/forgot-password` | Forgot Password | `POST /auth/forgot-password` | Email input → success message |
| `/reset-password` | Reset Password | `POST /auth/reset-password` | Token + new password form |

### Authenticated Pages (JWT required)

| Route | Page | Backend API | Description |
|---|---|---|---|
| `/dashboard` | User Dashboard | `GET /auth/me` | Overview: bookmarks count, alerts, recent activity |
| `/dashboard/profile` | Edit Profile | `GET /auth/me` / `PUT /users/me` | Edit name, phone, district, education |
| `/dashboard/bookmarks` | My Bookmarks | `GET /users/me/bookmarks` | List of bookmarked circulars with remove button |
| `/dashboard/alerts` | My Alerts | `GET /users/me/alerts` | List alert rules, create new, toggle active, delete |
| `/dashboard/alerts/new` | Create Alert | `POST /users/me/alerts` | Form: keyword, category, organization, education |

### Admin Pages (role = "admin")

| Route | Page | Backend API | Description |
|---|---|---|---|
| `/admin` | Admin Dashboard | `GET /admin/stats` | Stats cards, recent scrape logs, quick actions |
| `/admin/users` | User Management | `GET /admin/users` | Table of users, view details |
| `/admin/circulars` | Circular Management | `GET /circulars?status=all` | Admin CRUD for circulars |
| `/admin/circulars/new` | Create Circular | `POST /circulars` | Manual circular creation form |
| `/admin/circulars/[id]/edit` | Edit Circular | `PUT /circulars/:id` | Edit form |
| `/admin/scrape` | Scrape Control | `POST /admin/scrape/run` | Trigger button, recent logs table |

### Route Group Layout

```
app/
├── (public)/                  # public layout (nav + footer)
│   ├── layout.tsx
│   ├── page.tsx               # home
│   ├── circulars/
│   │   ├── page.tsx           # listing
│   │   └── [id]/page.tsx      # detail
│   ├── categories/
│   │   ├── page.tsx           # all categories
│   │   └── [slug]/page.tsx    # filtered
│   ├── organizations/page.tsx
│   ├── login/page.tsx
│   ├── register/page.tsx
│   ├── verify-email/page.tsx
│   ├── forgot-password/page.tsx
│   └── reset-password/page.tsx
│
├── (auth)/                    # auth-required layout (dashboard sidebar)
│   ├── layout.tsx
│   ├── dashboard/
│   │   ├── page.tsx           # overview
│   │   ├── profile/page.tsx
│   │   ├── bookmarks/page.tsx
│   │   ├── alerts/page.tsx
│   │   └── alerts/new/page.tsx
│   └── settings/page.tsx
│
└── (admin)/                   # admin-required layout
    ├── layout.tsx
    └── admin/
        ├── page.tsx           # dashboard stats
        ├── users/page.tsx
        ├── circulars/
        │   ├── page.tsx       # manage
        │   ├── new/page.tsx   # create
        │   └── [id]/edit/page.tsx
        └── scrape/page.tsx
```

---

## 3. Component Architecture

### Layout Components

```
components/
├── layout/
│   ├── PublicLayout.tsx        # Navbar (brand, search, login/avatar) + Footer
│   ├── Navbar.tsx              # Responsive nav with auth state
│   ├── Footer.tsx              # Links, about, contact
│   ├── DashboardLayout.tsx     # Sidebar + main content area
│   ├── AdminLayout.tsx         # Admin sidebar + main content
│   └── Sidebar.tsx             # Navigation links with active state
```

### Shared Components

```
components/
├── ui/                         # shadcn/ui primitives (Button, Input, Card, etc.)
├── shared/
│   ├── CircularCard.tsx         # Card used in lists (title, org, deadline, category badge)
│   ├── CircularCardSkeleton.tsx # Loading skeleton
│   ├── CircularDetail.tsx       # Full detail view (all fields)
│   ├── CategoryPills.tsx        # Horizontal scrollable category buttons
│   ├── SearchBar.tsx            # Search input with debounce + category filter
│   ├── Pagination.tsx           # Page numbers, prev/next
│   ├── FilterPanel.tsx          # Sidebar/modal with all filters
│   ├── BookmarkButton.tsx       # Heart icon toggle (optimistic update)
│   ├── DeadlineBadge.tsx        # Colored badge for deadline (green: far, yellow: soon, red: expired)
│   ├── EmptyState.tsx           # "No circulars found" with illustration
│   ├── ErrorState.tsx           # Error display with retry button
│   ├── LoadingSpinner.tsx
│   └── ThemeToggle.tsx          # Light/dark mode (use next-themes)
```

### Feature Components

```
components/
├── auth/
│   ├── LoginForm.tsx
│   ├── RegisterForm.tsx
│   ├── ForgotPasswordForm.tsx
│   ├── ResetPasswordForm.tsx
│   ├── AuthGuard.tsx            # Wrapper: redirect to login if no token
│   └── AdminGuard.tsx           # Wrapper: redirect to 403 if not admin
│
├── circulars/
│   ├── CircularFilters.tsx      # Category, status, date range, education, gender, sort
│   ├── CircularList.tsx         # Infinite scroll or paginated list
│   ├── CircularDetailPage.tsx   # Full page with all sections
│   ├── ApplyButton.tsx          # External link or Teletalk redirect
│   └── CircularForm.tsx         # Admin create/edit form (all 30+ fields)
│
├── bookmarks/
│   ├── BookmarkList.tsx         # With circular cards + remove buttons
│   └── BookmarkCount.tsx        # Badge in navbar showing count
│
├── alerts/
│   ├── AlertList.tsx            # List of user's alert rules
│   ├── AlertCard.tsx            # Single alert with toggle + delete
│   ├── AlertForm.tsx            # Create/edit alert form
│   └── AlertMatchPreview.tsx    # Preview matching circulars (if backend supports)
│
├── profile/
│   ├── ProfileForm.tsx          # Edit name, phone, district, education
│   └── ProfileCard.tsx          # Display-only view
│
└── admin/
    ├── StatsCards.tsx           # Total circulars, active, expired, users, scrape stats
    ├── UsersTable.tsx           # Sortable/filterable user table
    ├── ScrapeControl.tsx        # Trigger scrape button + recent scrape logs table
    └── CircularManageTable.tsx  # Admin CRUD table with featured toggle
```

---

## 4. Auth Flow (Detailed)

### Token Storage

| Token | Storage | Usage |
|---|---|---|
| `access_token` | **In-memory** (Zustand store) | `Authorization: Bearer <token>` header |
| `refresh_token` | **httpOnly cookie** (set by backend) | Sent automatically with every request |

### Auth Store (Zustand)

```typescript
// stores/auth-store.ts
interface AuthState {
  accessToken: string | null;
  user: UserProfile | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  
  login: (email: string, password: string) => Promise<void>;
  register: (data: RegisterInput) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
  fetchProfile: () => Promise<void>;
}
```

### Auth Flow Sequence

```
┌─────────────────────────────────────────────────────────────────┐
│ REGISTER                                                         │
│ Form → POST /auth/register → success → "check email" message     │
│ → Email arrives → click link → /verify-email?token=xxx           │
│ → GET /auth/verify-email?token=xxx → verified → redirect to login │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ LOGIN                                                            │
│ Form → POST /auth/login → { access_token, user }                 │
│ → Store access_token in Zustand                                  │
│ → refresh_token saved as httpOnly cookie automatically           │
│ → Redirect to /dashboard or previous page                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ AUTHENTICATED REQUEST                                            │
│ API client interceptor → attach Bearer token                     │
│ → If 401 → attempt POST /auth/refresh (cookie sent automatically) │
│ → If refresh succeeds → retry original request                   │
│ → If refresh fails → clear state → redirect to /login            │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ LOGOUT                                                           │
│ → POST /auth/logout (invalidates refresh token in Redis)         │
│ → Clear access_token from Zustand                                │
│ → Backend clears refresh_token cookie                            │
│ → Redirect to /                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### API Client Setup

```typescript
// lib/api.ts
class ApiClient {
  private accessToken: string | null = null;
  private refreshPromise: Promise<void> | null = null;

  setToken(token: string | null) { this.accessToken = token; }

  async fetch<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(this.accessToken ? { Authorization: `Bearer ${this.accessToken}` } : {}),
        ...options?.headers,
      },
      credentials: 'include', // sends httpOnly cookie for refresh
    });

    if (res.status === 401 && this.accessToken) {
      // Token expired — try refresh
      if (!this.refreshPromise) {
        this.refreshPromise = this.refreshToken();
      }
      await this.refreshPromise;
      this.refreshPromise = null;
      // Retry original request
      return this.fetch<T>(endpoint, options);
    }

    if (!res.ok) throw new ApiError(res.status, await res.json());
    return res.json();
  }

  private async refreshToken() {
    const res = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    });
    if (!res.ok) { this.setToken(null); throw new Error('Session expired'); }
    const data = await res.json();
    this.setToken(data.data.access_token);
  }
}
```

---

## 5. Data Fetching Strategy

### Pattern per page type

| Page Type | Strategy | Example |
|---|---|---|
| **Public listing** | Server Component + SSR | `/circulars` — fetch on server, render HTML |
| **Public detail** | Server Component + generateMetadata | `/circulars/[id]` — SEO metadata from API |
| **Search/filter** | Client Component + TanStack Query | Filters change → client-side refetch |
| **Authenticated pages** | Client Component + TanStack Query | `/dashboard/bookmarks` — fetch after mount |
| **Forms** | React Hook Form + useMutation | Register, login, profile edit |
| **Optimistic updates** | useMutation with onMutate | Bookmark toggle |

### Example: Circular listing page

```typescript
// app/(public)/circulars/page.tsx (Server Component — initial load)
export default async function CircularsPage({ searchParams }) {
  const params = new URLSearchParams(searchParams);
  const data = await apiClient.fetch(`/circulars?${params}`);
  
  return (
    <CircularList initialData={data}>
      <FilterPanel />
      <CircularGrid circulars={data.data.circulars} />
      <Pagination pagination={data.data.pagination} />
    </CircularList>
  );
}

// components/circulars/CircularList.tsx (Client Component — interactive)
'use client';
export function CircularList({ initialData }) {
  const [filters, setFilters] = useState(parseFiltersFromURL());
  
  const { data, isLoading } = useQuery({
    queryKey: ['circulars', filters],
    queryFn: () => apiClient.fetch(`/circulars?${new URLSearchParams(filters)}`),
    initialData, // SSR data as initial cache
  });
  
  return (/* ... */);
}
```

---

## 6. Feature Deep Dives

### 6.1 Home Page (`/`)

```
┌─────────────────────────────────────────────┐
│  [Navbar: Logo · Categories · Search · Login]│
├─────────────────────────────────────────────┤
│                                              │
│    🇧🇩 বাংলাদেশ সরকারি চাকরির সার্কুলার         │
│      Find government jobs in one place        │
│                                              │
│   ┌─────────────────────────────────────┐    │
│   │ 🔍 Search jobs...          [Filter] │    │
│   └─────────────────────────────────────┘    │
│                                              │
│   [⭐BCS][🏦Bank][🛡Defense][👮Police]...     │  ← CategoryPills
│                                              │
│   📌 Featured Circulars                      │
│   ┌────────┐ ┌────────┐ ┌────────┐          │
│   │ Card 1 │ │ Card 2 │ │ Card 3 │          │  ← CircularCard
│   └────────┘ └────────┘ └────────┘          │
│                                              │
│   📋 Latest Circulars                        │
│   List of cards with pagination              │
│                                              │
│   📊 Stats Bar                               │
│   500+ Active · 50 Organizations · Updated Daily │
│                                              │
└─────────────────────────────────────────────┘
```

### 6.2 Circular Listing (`/circulars`)

**Filters (sticky sidebar or top bar):**
- Category dropdown (from `GET /categories`)
- Status: Active | Expired | All
- Education level (SSC, HSC, Degree, Masters)
- Gender (Male, Female, Both)
- Date range (deadline from/to)
- Sort: Latest | Deadline (soonest) | Most viewed

**Card design (CircularCard):**
```
┌──────────────────────────────────────────┐
│ 🏦 Bank Jobs                    ⏰ 15 days │  ← DeadlineBadge
│                                              │
│ Assistant Director                           │
│ Bangladesh Bank · Dhaka                      │
│                                              │
│ ⏳ Vacancy: 50    💰 Tk. 35,500 (Grade-6)   │
│ 📅 Published: 01 Jun  ·  Deadline: 30 Jun   │
│                                              │
│ [Apply via Teletalk]              [🤍 Bookmark]│
└──────────────────────────────────────────┘
```

### 6.3 Circular Detail (`/circulars/[id]`)

**Sections:**
1. Title + Organization (with logo if available)
2. Category badge
3. Key Info grid: Vacancy, Salary, Deadline, Location, Job Type, Gender
4. Age requirements
5. Education requirements (detailed)
6. Experience requirements
7. Description (rich text / markdown)
8. Requirements (rich text / markdown)
9. Application instructions (apply_via, teletalk_code, apply_url)
10. Important dates (published, deadline, exam)
11. Circular image (if circular_image_url)
12. Circular PDF (if circular_pdf_url — link to MinIO)
13. Related circulars (same category/organization)
14. Bookmark button (if authenticated)
15. Share buttons

### 6.4 Bookmarks (`/dashboard/bookmarks`)

```typescript
// Optimistic toggle
const toggleBookmark = useMutation({
  mutationFn: ({ id, isBookmarked }) =>
    isBookmarked
      ? apiClient.fetch(`/users/me/bookmarks/${id}`, { method: 'DELETE' })
      : apiClient.fetch(`/users/me/bookmarks/${id}`, { method: 'POST' }),
  onMutate: async ({ id }) => {
    // Optimistically toggle UI
    queryClient.setQueryData(['bookmarks'], (old) => /* toggle */);
  },
  onError: () => {
    // Rollback on failure
    queryClient.invalidateQueries(['bookmarks']);
  },
});
```

### 6.5 Email Alerts (`/dashboard/alerts`)

**Alert rule fields (from `alerts` table):**
- `keyword` — free text to match against circular title/description
- `category_id` — dropdown from categories
- `organization_id` — searchable dropdown from organizations
- `education_level` — dropdown (SSC, HSC, Degree, Masters)
- `is_active` — toggle on/off

**Note:** Email alert delivery (cron job) is not yet implemented in the backend. Build the frontend UI now with a tooltip: *"Alerts will send you an email when matching circulars are found. This feature is coming soon."*

### 6.6 Search

**Search bar behavior:**
- Debounced input (300ms)
- Searches `GET /circulars/search?q=keyword`
- Shows suggestions dropdown (if backend supports)
- Updates URL query params (so search is shareable)

---

## 7. Admin Dashboard

### Admin Login

Admins use the same `/login` page. The role comes from the JWT claims (`role: "admin"`). The `AdminGuard` component checks this and redirects non-admins to a 403 page.

### Admin Pages Detail

**`/admin` — Dashboard:**
```
┌──────────────────────────────────────────────────────────┐
│ Admin Dashboard                                          │
│                                                          │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│ │ Active   │ │ Total    │ │ Users    │ │ Last     │    │
│ │ 345      │ │ 1,250    │ │ 89       │ │ Scrape   │    │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
│                                                          │
│ Recent Scrape Logs                          [Run Scrape] │
│ ┌──────────────────────────────────────────────────────┐ │
│ │ bdjobs    │ success │ 120 fetched │ 5 new │ 2m ago   │ │
│ │ teletalk  │ success │ 80 fetched  │ 3 new │ 8m ago   │ │
│ │ bdjobs    │ failed  │ -           │ -     │ 6h ago   │ │
│ └──────────────────────────────────────────────────────┘ │
│                                                          │
│ Recent Users                                             │
│ ┌──────────────────────────────────────────────────────┐ │
│ │ Name    │ Email           │ Role  │ Verified │ Date  │ │
│ │ Fuad    │ fuad@email.com  │ user  │ ✅       │ Today │ │
│ └──────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

**`/admin/circulars` — Management table:**
- Paginated table with all circulars
- Status column with colored badge (active/expired/closed)
- Featured toggle (inline PATCH)
- Edit / Delete action buttons
- "Add New" button → `/admin/circulars/new`

**`/admin/circulars/new` & `edit` — Form:**
- Full form matching all 30+ circular fields
- Organization autocomplete/search
- Category dropdown
- Date pickers (published, deadline, exam)
- Rich text editor for description/requirements (TipTap or similar)
- Upload circular image/PDF → MinIO

---

## 8. File / Asset Handling (MinIO)

### Architecture

```
┌──────────┐    Upload      ┌──────────┐   Store    ┌──────────┐
│ Frontend │ ─────────────> │ Go API   │ ────────> │  MinIO   │
│ (Next.js)│ <───────────── │ :8080    │ <──────── │  :9000   │
└──────────┘   Return URL   └──────────┘  Get URL   └──────────┘
```

**Important:** Do NOT expose MinIO directly to the browser. All uploads go through the Go API (authentication, validation, virus scanning later). The API returns the MinIO URL, and the frontend displays it.

### Upload Flow (when backend implements it)

1. Admin clicks "Upload Image/PDF" in circular form
2. File selected via `<input type="file">`
3. `POST /admin/upload` with `multipart/form-data` + JWT
4. Backend stores file in MinIO, returns URL
5. Frontend stores URL in circular form state
6. On submit, URL is sent with `POST /circulars`

### Displaying Assets

```tsx
// Circular image
<Image src={circular.circular_image_url} alt={circular.title} />

// Circular PDF
<a href={circular.circular_pdf_url} target="_blank">
  📄 Download Circular PDF
</a>
```

---

## 9. SEO & Performance

### Metadata per page

```typescript
// app/(public)/circulars/[id]/page.tsx
export async function generateMetadata({ params }) {
  const circular = await fetchCircular(params.id);
  return {
    title: `${circular.title} — BD Govt Jobs`,
    description: `${circular.organization_name} — Vacancy: ${circular.vacancy}. Apply by ${circular.application_deadline}.`,
    openGraph: {
      title: circular.title,
      description: `Apply for ${circular.title} at ${circular.organization_name}`,
      images: [circular.circular_image_url],
    },
  };
}
```

### Performance targets

| Metric | Target |
|---|---|
| Lighthouse Score | 90+ |
| LCP (Largest Contentful Paint) | < 2.5s |
| FID (First Input Delay) | < 100ms |
| First Load JS | < 150KB |
| Circular list TTI | < 2s (with SSG/ISR) |

### Strategies

- **ISR (Incremental Static Regeneration)** for circular listing pages (revalidate every 5 minutes)
- **SSR** for circular detail (always fresh — deadlines matter)
- **SSG** for static pages (about, categories, faq)
- **Loading skeletons** everywhere (CircularCardSkeleton for list items)
- **Image optimization** via `next/image` with MinIO domains configured

---

## 10. Internationalization (Bangla)

### Approach

Bangladeshi government job seekers are primarily Bangla speakers. Support both **English** and **Bangla** (বাংলা).

- **`next-intl`** for routing-based i18n (`/en/...`, `/bn/...`)
- Default: Bangla
- Detect from `Accept-Language` header, with manual toggle

### What gets translated

| Content | Source |
|---|---|
| UI labels | Static translation files (JSON) |
| Category names | Already in DB (`name` + `name_bn`) |
| Circular titles | Already in DB (`title` + `title_bn`) |
| Organization names | Already in DB (`name` + `name_bn`) |
| Error messages | Translation files |

### Translation file structure

```
messages/
├── en.json
│   { "home.search": "Search jobs...", "home.featured": "Featured Circulars", ... }
├── bn.json
│   { "home.search": "চাকরি খুঁজুন...", "home.featured": "বিশেষ সার্কুলার", ... }
```

---

## 11. Environment Variables

```env
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_MINIO_CONSOLE_URL=http://localhost:9001
NEXT_PUBLIC_SITE_NAME=BD Govt Jobs
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000

# Production
# NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
# NEXT_PUBLIC_FRONTEND_URL=https://yourdomain.com
```

---

## 12. Project Structure

```
job-circular-client/
├── app/
│   ├── (public)/
│   │   ├── layout.tsx
│   │   ├── page.tsx                    # home
│   │   ├── circulars/
│   │   │   ├── page.tsx                # listing
│   │   │   └── [id]/page.tsx           # detail
│   │   ├── categories/
│   │   │   ├── page.tsx
│   │   │   └── [slug]/page.tsx
│   │   ├── organizations/page.tsx
│   │   ├── login/page.tsx
│   │   ├── register/page.tsx
│   │   ├── verify-email/page.tsx
│   │   ├── forgot-password/page.tsx
│   │   └── reset-password/page.tsx
│   │
│   ├── (auth)/
│   │   ├── layout.tsx
│   │   └── dashboard/
│   │       ├── page.tsx
│   │       ├── profile/page.tsx
│   │       ├── bookmarks/page.tsx
│   │       ├── alerts/
│   │       │   ├── page.tsx
│   │       │   └── new/page.tsx
│   │       └── settings/page.tsx
│   │
│   ├── (admin)/
│   │   ├── layout.tsx
│   │   └── admin/
│   │       ├── page.tsx
│   │       ├── users/page.tsx
│   │       ├── circulars/
│   │       │   ├── page.tsx
│   │       │   ├── new/page.tsx
│   │       │   └── [id]/edit/page.tsx
│   │       └── scrape/page.tsx
│   │
│   ├── error.tsx
│   ├── not-found.tsx
│   └── globals.css
│
├── components/
│   ├── ui/                    # shadcn/ui generated components
│   ├── layout/                # Navbar, Footer, Sidebar, Layouts
│   ├── shared/                # CircularCard, SearchBar, Pagination, etc.
│   ├── auth/                  # LoginForm, RegisterForm, AuthGuard, AdminGuard
│   ├── circulars/             # CircularFilters, CircularList, etc.
│   ├── bookmarks/
│   ├── alerts/
│   ├── profile/
│   └── admin/                 # StatsCards, UsersTable, ScrapeControl
│
├── lib/
│   ├── api.ts                 # ApiClient with token refresh
│   ├── validators.ts          # Zod schemas
│   └── utils.ts               # Date formatting, URL helpers
│
├── stores/
│   └── auth-store.ts          # Zustand auth state
│
├── hooks/
│   ├── use-auth.ts            # Auth convenience hook
│   ├── use-bookmarks.ts       # Bookmark operations
│   └── use-debounce.ts
│
├── types/
│   ├── circular.ts
│   ├── user.ts
│   ├── category.ts
│   ├── organization.ts
│   └── api.ts                 # API response types
│
├── messages/                  # i18n translation files
│   ├── en.json
│   └── bn.json
│
├── public/                    # Static assets (favicon, og-image, etc.)
├── .env.local
├── .env.local.example
├── next.config.js
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

---

## 13. Deployment

### Option A: Same server as backend (Docker)

```dockerfile
# Dockerfile (frontend)
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./
EXPOSE 3000
CMD ["pnpm", "start"]
```

Add to `docker-compose.prod.yml`:
```yaml
frontend:
  build: ../job-circular-client
  env_file: ../job-circular-client/.env.local
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.frontend.rule=Host(`yourdomain.com`)"
    - "traefik.http.routers.frontend.entrypoints=websecure"
    - "traefik.http.routers.frontend.tls.certresolver=le"
    - "traefik.http.services.frontend.loadbalancer.server.port=3000"
```

### Option B: Vercel (recommended for Next.js)

```bash
vercel --prod
```

Set env vars in Vercel dashboard:
```
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
```

---

## 14. Implementation Phases

### Phase 1: Foundation (Week 1)

- [ ] Initialize Next.js project with TypeScript, Tailwind, shadcn/ui
- [ ] Set up project structure (folders, aliases, linting)
- [ ] Create `ApiClient` in `lib/api.ts` with token refresh
- [ ] Implement Zustand auth store
- [ ] Build `PublicLayout` (Navbar + Footer)
- [ ] Build auth guard components (`AuthGuard`, `AdminGuard`)
- [ ] Set up i18n with `next-intl` (en + bn)
- [ ] Create type definitions from API response shapes
- [ ] Add `.env.local.example`

### Phase 2: Auth (Week 1-2)

- [ ] Login page + form
- [ ] Register page + form (with optional phone, district, education)
- [ ] Email verification page
- [ ] Forgot password → Reset password flow
- [ ] Auth state persistence (refresh on app load)
- [ ] 401 interception + automatic token refresh
- [ ] Logout (clear state + call API)

### Phase 3: Core Pages (Week 2-3)

- [ ] Home page (hero, search bar, category pills, featured circulars)
- [ ] Circular listing page (with all filters, pagination)
- [ ] Circular detail page (all 14 sections)
- [ ] Category pages (grid + filtered listing)
- [ ] Organization listing page
- [ ] Search functionality (debounced, URL-based)
- [ ] Loading skeletons for all card/list components
- [ ] Error states + empty states

### Phase 4: User Features (Week 3-4)

- [ ] Dashboard layout (sidebar + overview)
- [ ] Profile page (view + edit)
- [ ] Bookmarks list with optimistic toggle
- [ ] Alerts list (view all, toggle active, delete)
- [ ] Create alert form
- [ ] Bookmark button on circular cards + detail page
- [ ] Bookmark count badge in navbar

### Phase 5: Admin (Week 4-5)

- [ ] Admin layout + route guard
- [ ] Admin dashboard with stats cards
- [ ] Circular management (table + create + edit + delete + featured toggle)
- [ ] User management table
- [ ] Scrape control (trigger button + logs table)
- [ ] Circular image/PDF upload (after backend implements MinIO upload endpoint)

### Phase 6: Polish (Week 5-6)

- [ ] Responsive testing (mobile-first — most users are on mobile)
- [ ] Dark mode support
- [ ] SEO metadata for all pages
- [ ] OG images for circular detail pages
- [ ] Accessibility audit (WCAG 2.1 AA)
- [ ] Performance optimization (bundle analysis, image optimization)
- [ ] Bangla translation completion
- [ ] Browser testing (Chrome, Firefox, Safari, mobile browsers)
- [ ] PWA support (offline circular list, push notifications for alerts)

### Phase 7: Deploy & Monitor (Week 6)

- [ ] Deploy to Vercel or Docker on DO server
- [ ] Set up CORS (configure `FRONTEND_URL` in backend .env)
- [ ] Configure SSL/TLS
- [ ] Add analytics
- [ ] Set up error monitoring (Sentry or similar)
- [ ] Load testing against backend API

---

## API Endpoint → Frontend Page Mapping (Complete)

| API Endpoint | Frontend Page | Status |
|---|---|---|
| `GET /health` | Used by monitoring / Vercel health checks | Backend ✅ |
| `POST /auth/register` | `/register` | Backend ✅ |
| `POST /auth/login` | `/login` | Backend ✅ |
| `GET /auth/verify-email?token=` | `/verify-email` | Backend ✅ |
| `POST /auth/forgot-password` | `/forgot-password` | Backend ✅ |
| `POST /auth/reset-password` | `/reset-password` | Backend ✅ |
| `POST /auth/refresh` | ApiClient interceptor | Backend ✅ |
| `POST /auth/logout` | Navbar logout button | Backend ✅ |
| `GET /auth/me` | `/dashboard`, Navbar avatar | Backend ✅ |
| `GET /circulars` | `/circulars`, `/[slug]`, homepage | Backend ❌ |
| `GET /circulars/:id` | `/circulars/[id]` | Backend ❌ |
| `GET /circulars/search?q=` | Search bar (all pages) | Backend ❌ |
| `GET /circulars/featured` | Homepage featured section | Backend ❌ |
| `GET /categories` | `/categories`, filter dropdowns | Backend ❌ |
| `GET /organizations` | `/organizations`, alert form | Backend ❌ |
| `GET /users/me` | `/dashboard/profile` | Backend ❌ |
| `PUT /users/me` | Profile edit form | Backend ❌ |
| `GET /users/me/bookmarks` | `/dashboard/bookmarks` | Backend ❌ |
| `POST /users/me/bookmarks/:id` | BookmarkButton component | Backend ❌ |
| `DELETE /users/me/bookmarks/:id` | BookmarkButton component | Backend ❌ |
| `GET /users/me/alerts` | `/dashboard/alerts` | Backend ❌ |
| `POST /users/me/alerts` | `/dashboard/alerts/new` | Backend ❌ |
| `DELETE /users/me/alerts/:id` | AlertCard delete button | Backend ❌ |
| `POST /circulars` | `/admin/circulars/new` | Backend ❌ |
| `PUT /circulars/:id` | `/admin/circulars/[id]/edit` | Backend ❌ |
| `DELETE /circulars/:id` | Admin manage table | Backend ❌ |
| `PATCH /circulars/:id/feature` | Featured toggle | Backend ❌ |
| `GET /admin/stats` | `/admin` dashboard | Backend ❌ |
| `GET /admin/users` | `/admin/users` | Backend ❌ |
| `POST /admin/scrape/run` | `/admin/scrape` button | Backend ❌ |
| `GET /admin/scrape/logs` | `/admin/scrape` logs table | Backend ❌ |

---

*Plan last updated: June 2026*
