# JobCirculer — Frontend Documentation

> **Stack:** Vanilla HTML/CSS/JS (no framework)  
> **Served by:** Go backend via `http.FileServer`  
> **Font:** Inter (Google Fonts)

---

## 1. File Structure

```
frontend/
├── index.html          # Single-page shell: 4 views + auth panel + navbar
├── css/
│   └── style.css       # All styles (~500 lines, dark theme, responsive)
└── js/
    ├── data.js          # Demo data (15 circulars, 10 categories, 8 orgs)
    └── app.js           # App logic: routing, rendering, auth, bookmarks
```

---

## 2. Architecture

The frontend is a **single-page application** with client-side view switching. There is no router library — views are toggled by adding/removing the `.active` CSS class on `<div>` containers.

### Views

| View | DOM ID | Access | Purpose |
|---|---|---|---|
| Homepage | `#home-view` | Public (default) | Hero section, featured circulars, how-it-works |
| Browse Jobs | `#circulars-view` | Public | Search, category filter, sort, paginated grid |
| Job Detail | `#circular-detail-view` | Public | Full circular info, apply link, bookmark |
| Dashboard | `#dashboard-view` | Auth required | User bookmarks, alerts |

### Persistent Elements (always mounted)

| Element | Purpose |
|---|---|
| `.navbar` | Fixed top bar with logo, nav links, auth buttons / user menu |
| `#auth-overlay` | Slide-in panel (right side) with login/register forms |

---

## 3. Navigation System

Navigation is handled by the `navigateTo(view, circularId?)` function in `app.js`. It:

1. Hides all views (removes `.active`)
2. Shows the target view
3. Highlights the correct nav link
4. Triggers data rendering for the target view

Supported routes:
- `navigateTo('home')` — renders featured circulars
- `navigateTo('circulars')` — applies filters + renders list
- `navigateTo('detail', 'c001')` — renders full circular detail
- `navigateTo('dashboard')` — redirects to home if not logged in

---

## 4. Data Layer

### Demo Data (`data.js`)

All data is hardcoded since the backend circular endpoints are not yet implemented. Contains:

| Variable | Description |
|---|---|
| `CATEGORIES` | 10 categories with Bengali names, slugs, and icons |
| `ORGANIZATIONS` | 8 organizations with types |
| `DEMO_CIRCULARS` | 15 circulars with full details (title, org, dates, salary, description, etc.) |
| `CATEGORY_FILTERS` | "All" + the 10 categories (for filter pills) |
| `SORT_OPTIONS` | Newest first, deadline soonest, most viewed |

### Real API Integration (future)

When backend endpoints are ready, replace demo data calls with:
```js
GET /api/v1/circulars              → list (with query params for filter/search/pagination)
GET /api/v1/circulars/:id          → detail
GET /api/v1/categories             → category list
POST /api/v1/users/me/bookmarks/:id  → save bookmark (auth required)
DELETE /api/v1/users/me/bookmarks/:id → remove bookmark
```

---

## 5. State Management

No state library — uses plain variables and `localStorage`.

| Storage | Key | Contents |
|---|---|---|
| `localStorage` | `access_token` | JWT token string |
| `localStorage` | `user` | User object (name, email, etc.) |
| `localStorage` | `bookmarks` | Array of circular IDs |
| Memory (app.js) | `currentView` | Active view name |
| Memory (app.js) | `currentPage` | Current pagination page |
| Memory (app.js) | `currentFilters` | `{ category, search, sort }` object |
| Memory (app.js) | `currentCircularId` | ID for detail view |

---

## 6. Auth Flow

```
Guest user visits site
  → Homepage loads (public)
  → Can browse/search/view circulars (all public)

User clicks bookmark → requireAuth() checks
  → If NOT logged in → opens auth slide panel
  → If logged in → toggles bookmark in localStorage

Auth panel:
  → Login tab (email + password)
  → Register tab (name + email + password)
  → Calls POST /api/v1/auth/login or /register
  → On success: stores token + user, closes panel
  → Register auto-logs in after success

Logout:
  → Calls POST /api/v1/auth/logout
  → Clears localStorage
  → Returns to homepage
```

---

## 7. Styling Approach

- **Design tokens** in CSS custom properties (`:root`) — indigo accent, dark slate palette
- **Dark theme** — slate-900 background, slate-800 cards, indigo-600 primary
- **No CSS framework** — hand-written, consistent spacing/color system
- **Responsive** — 3-col → 2-col → 1-col grid breakpoints at 900px and 768px
- **Smooth transitions** — hover lifts, slide-in auth panel, fade-in forms

Key design tokens:
```css
--primary: #4F46E5;      /* Indigo-600 */
--bg-main: #0F172A;      /* Slate-900 */
--bg-card: #1E293B;      /* Slate-800 */
--text-primary: #F8FAFC; /* Slate-50 */
--text-secondary: #94A3B8; /* Slate-400 */
--border: #334155;       /* Slate-700 */
```

---

## 8. TODO

### Immediate
- [ ] Replace demo data with real API calls when backend circular endpoints are ready
- [ ] Add loading skeletons / spinners for API fetches
- [ ] Implement bookmark persistence via API (currently localStorage only)
- [ ] Add "Forgot Password" flow in auth panel
- [ ] Add mobile hamburger menu for nav links
- [ ] Add toast notifications for actions (bookmark added, login success, errors)

### Short Term
- [ ] **Circular list enhancements:**
  - [ ] Filter by organization
  - [ ] Filter by education level
  - [ ] Filter by deadline range (date inputs)
  - [ ] Filter by location/district
- [ ] **Circular detail enhancements:**
  - [ ] Show related circulars at bottom
  - [ ] Share button (copy link)
  - [ ] Show circular image/PDF if available
- [ ] **Dashboard:**
  - [ ] Alert creation UI (keyword + category + organization)
  - [ ] Alert list with toggle active/inactive
  - [ ] Profile editing section
- [ ] Accessibility improvements (keyboard nav, ARIA labels, focus management)
- [ ] Error boundary / global error handling

### Mid Term
- [ ] Migrate to a framework (React/Vue) for better component reusability
- [ ] Add `react-router` or `vue-router` for proper client-side routing
- [ ] Server-side rendering for SEO (circular pages should be indexable)
- [ ] Offline support with Service Workers (cache circulars for offline browsing)
- [ ] Multi-language support (English + Bengali toggle)
- [ ] Dark/light theme toggle

---

## 9. Future Plans

### Phase 1: Polish & Launch (Next)
- Connect frontend to real backend API
- Email verification flow in auth panel
- Password reset flow
- SEO meta tags for circular pages
- Basic analytics (page views, popular circulars)

### Phase 2: Power Features
- **Smart Alerts:** Users set criteria (keywords, categories, education level), get email/push notifications when matching circulars are published
- **Application Tracker:** Users mark circulars as "Applied", "Shortlisted", "Interview", "Selected" — visual pipeline
- **Auto-fill:** Pre-fill Teletalk/online application forms from saved profile data
- **Circular Comparison:** Side-by-side comparison of up to 3 circulars
- **Deadline Calendar:** Calendar view showing all upcoming deadlines

### Phase 3: Community & Growth
- **Discussion forum** per circular — users discuss exam prep, share tips
- **Exam resources** — previous year questions, syllabus links per circular
- **Success stories** — user-submitted success stories with optional anonymity
- **Referral system** — invite friends, earn premium features

### Phase 4: Scale & Automate
- **AI-powered recommendations** — suggest circulars based on user profile, browsing history, and bookmarked patterns
- **Auto-apply bot** — with user consent, automatically fill and submit applications on Teletalk/online portals
- **Circular authenticity verification** — crowd-sourced flagging of fake/scam circulars
- **Mobile apps** — React Native / Flutter for iOS and Android
- **API monetization** — paid API access for third-party job boards or HR platforms

### Phase 5: Beyond Govt Jobs
- Expand to private sector jobs
- International job listings (for Bangladeshis abroad)
- Freelance/gig marketplace
- Employer dashboard for posting jobs directly
- Recruitment CRM for organizations
