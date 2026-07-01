// ═══════════════════════════════════════════════════════════════════════════════
//  JobCirculer — app.js  (API-backed)
// ═══════════════════════════════════════════════════════════════════════════════

// ── Configuration ──────────────────────────────────────────────────────────────
var API_URL = "/api/v1";
var ITEMS_PER_PAGE = 6;
var currentFilters = { search: "", sort: "published_desc", category: "" };
var currentPage = 1;
var currentCircularId = null;

// ── Auth state ─────────────────────────────────────────────────────────────────
var token = localStorage.getItem("access_token") || null;
var user = JSON.parse(localStorage.getItem("user") || "null");

// ── API helpers ────────────────────────────────────────────────────────────────
function apiGet(path) {
  var headers = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = "Bearer " + token;
  return fetch(API_URL + path, {
    headers: headers,
    credentials: "include",
  }).then(function (res) {
    return res.json().then(function (data) {
      return { ok: res.ok, status: res.status, body: data };
    });
  });
}

function apiPost(path, body) {
  var headers = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = "Bearer " + token;
  return fetch(API_URL + path, {
    method: "POST",
    headers: headers,
    body: body ? JSON.stringify(body) : undefined,
    credentials: "include",
  }).then(function (res) {
    return res.json().then(function (data) {
      return { ok: res.ok, status: res.status, body: data };
    });
  });
}

function apiDelete(path) {
  var headers = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = "Bearer " + token;
  return fetch(API_URL + path, {
    method: "DELETE",
    headers: headers,
    credentials: "include",
  }).then(function (res) {
    return res.json().then(function (data) {
      return { ok: res.ok, status: res.status, body: data };
    });
  });
}

function apiPatch(path, body) {
  var headers = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = "Bearer " + token;
  return fetch(API_URL + path, {
    method: "PATCH",
    headers: headers,
    body: body ? JSON.stringify(body) : undefined,
    credentials: "include",
  }).then(function (res) {
    return res.json().then(function (data) {
      return { ok: res.ok, status: res.status, body: data };
    });
  });
}

// Init — set up event delegation

document.addEventListener("DOMContentLoaded", function () {
  setAuthUI();
  document.addEventListener("click", handleDynamicClick);
});

// ═══════════════════════════════════════════════════════════════════════════════
//  Event Delegation  (bookmarks, pagination, cards, theme)
// ═══════════════════════════════════════════════════════════════════════════════
function handleDynamicClick(evt) {
  // Theme toggle
  var themeBtn = evt.target.closest("#theme-toggle");
  if (themeBtn) {
    toggleTheme();
    return;
  }
  // Bookmark button
  var bmBtn = evt.target.closest("[data-bookmark]");
  if (bmBtn) {
    evt.preventDefault();
    toggleBookmark(bmBtn.dataset.bookmark);
    return;
  }
  // Page button
  var pgBtn = evt.target.closest("[data-page]");
  if (pgBtn) {
    evt.preventDefault();
    currentPage = parseInt(pgBtn.dataset.page);
    renderCirculars();
    return;
  }
  // Card title → detail
  var cardTitle = evt.target.closest("[data-circular-id]");
  if (cardTitle) {
    evt.preventDefault();
    window.currentCircularId = cardTitle.dataset.circularId;
    htmx.ajax("GET", "/pages/circular-detail.html", {
      target: "#main-content",
      swap: "innerHTML",
    });
    return;
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
//  HTMX response interceptors
// ═══════════════════════════════════════════════════════════════════════════════

// Token refresh on 401
document.body.addEventListener("htmx:responseError", function (evt) {
  if (evt.detail.xhr.status === 401) {
    apiPost("/auth/refresh")
      .then(function (res) {
        if (
          res.ok &&
          res.body.success &&
          res.body.data &&
          res.body.data.access_token
        ) {
          token = res.body.data.access_token;
          user = res.body.data.user;
          saveAuth();
          setAuthUI();
        } else {
          clearAuth();
        }
      })
      .catch(function () {
        clearAuth();
      });
  }
});

// Auth form handlers — called directly from button onclick

function handleLoginFromForm(type) {
  var errorEl, email, password, name;

  if (type === "login") {
    email = document.getElementById("login-email");
    password = document.getElementById("login-password");
    errorEl = document.getElementById("login-error");
  } else {
    name = document.getElementById("register-name");
    email = document.getElementById("register-email");
    password = document.getElementById("register-password");
    errorEl = document.getElementById("register-error");
  }

  var emailVal = email ? email.value.trim() : "";
  var passwordVal = password ? password.value : "";
  var nameVal = name ? name.value.trim() : "";

  // Validation
  if (type === "register" && !nameVal) {
    if (errorEl) {
      errorEl.textContent = "All fields are required";
      errorEl.style.color = "var(--error)";
    }
    return;
  }
  if (!emailVal || !passwordVal) {
    if (errorEl) {
      errorEl.textContent = "Email and password are required";
      errorEl.style.color = "var(--error)";
    }
    return;
  }
  if (type === "register" && passwordVal.length < 6) {
    if (errorEl) {
      errorEl.textContent = "Password must be at least 6 characters";
      errorEl.style.color = "var(--error)";
    }
    return;
  }

  if (errorEl) errorEl.textContent = "Signing in...";

  if (type === "login") {
    apiPost("/auth/login", { email: emailVal, password: passwordVal })
      .then(function (res) {
        if (
          res.ok &&
          res.body.success &&
          res.body.data &&
          res.body.data.access_token
        ) {
          token = res.body.data.access_token;
          user = res.body.data.user;
          saveAuth();
          setAuthUI();
          fetchBookmarks();
          closeAuthPanel();
        } else {
          if (errorEl) {
            errorEl.textContent =
              (res.body && res.body.error) || "Login failed";
            errorEl.style.color = "var(--error)";
          }
        }
      })
      .catch(function () {
        if (errorEl) {
          errorEl.textContent = "Network error. Is the server running?";
          errorEl.style.color = "var(--error)";
        }
      });
  } else {
    apiPost("/auth/register", {
      name: nameVal,
      email: emailVal,
      password: passwordVal,
    })
      .then(function (res) {
        if (res.ok && res.body.success) {
          // Auto-login after registration
          apiPost("/auth/login", {
            email: emailVal,
            password: passwordVal,
          }).then(function (loginRes) {
            if (
              loginRes.ok &&
              loginRes.body.success &&
              loginRes.body.data &&
              loginRes.body.data.access_token
            ) {
              token = loginRes.body.data.access_token;
              user = loginRes.body.data.user;
              saveAuth();
              setAuthUI();
              fetchBookmarks();
              closeAuthPanel();
            } else {
              if (errorEl) {
                errorEl.textContent = "Account created! Please sign in.";
                errorEl.style.color = "var(--success)";
              }
              switchTab("login");
            }
          });
        } else {
          if (errorEl) {
            errorEl.textContent =
              (res.body && res.body.error) || "Registration failed";
            errorEl.style.color = "var(--error)";
          }
        }
      })
      .catch(function () {
        if (errorEl) {
          errorEl.textContent = "Network error. Is the server running?";
          errorEl.style.color = "var(--error)";
        }
      });
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Theme
// ═══════════════════════════════════════════════════════════════════════════════
function toggleTheme() {
  var html = document.documentElement;
  var current = html.getAttribute("data-theme") || "light";
  var next = current === "dark" ? "light" : "dark";
  html.setAttribute("data-theme", next);
  localStorage.setItem("theme", next);
  updateThemeIcons(next);
}

function updateThemeIcons(theme) {
  var btns = document.querySelectorAll("#theme-toggle");
  btns.forEach(function (btn) {
    btn.textContent = theme === "dark" ? "🌙" : "☀️";
    btn.setAttribute(
      "title",
      theme === "dark" ? "Switch to light" : "Switch to dark",
    );
    btn.setAttribute(
      "aria-label",
      theme === "dark" ? "Switch to light mode" : "Switch to dark mode",
    );
  });
}

// Keep theme icons in sync after HTMX swaps the navbar
document.body.addEventListener("htmx:afterSettle", function () {
  var theme = document.documentElement.getAttribute("data-theme") || "light";
  updateThemeIcons(theme);
});

// ═══════════════════════════════════════════════════════════════════════════════
//  Auth UI
// ═══════════════════════════════════════════════════════════════════════════════

function openAuthPanel() {
  var o = document.getElementById("auth-overlay");
  if (o) o.classList.add("open");
}

function closeAuthPanel() {
  var o = document.getElementById("auth-overlay");
  if (o) o.classList.remove("open");
}

function switchTab(tab) {
  var loginF = document.getElementById("login-form"),
    regF = document.getElementById("register-form");
  var tabL = document.getElementById("tab-login"),
    tabR = document.getElementById("tab-register");
  var lErr = document.getElementById("login-error"),
    rErr = document.getElementById("register-error");
  if (lErr) lErr.textContent = "";
  if (rErr) rErr.textContent = "";
  if (tab === "login") {
    loginF.classList.add("active");
    regF.classList.remove("active");
    tabL.classList.add("active");
    tabR.classList.remove("active");
  } else {
    regF.classList.add("active");
    loginF.classList.remove("active");
    tabR.classList.add("active");
    tabL.classList.remove("active");
  }
}

function setAuthUI() {
  var gb = document.getElementById("nav-auth-btns"),
    ud = document.getElementById("nav-user");
  var nd = document.getElementById("user-name-display"),
    dl = document.getElementById("nav-dashboard");
  var al = document.getElementById("nav-admin");

  if (token && user) {
    if (gb) gb.style.display = "none";
    if (ud) ud.style.display = "flex";
    if (nd) nd.textContent = user.name;
    if (dl) dl.style.display = "inline";
    if (al) al.style.display = user.role === "admin" ? "inline" : "none";
  } else {
    if (gb) gb.style.display = "flex";
    if (ud) ud.style.display = "none";
    if (dl) dl.style.display = "none";
    if (al) al.style.display = "none";
  }
}

function saveAuth() {
  localStorage.setItem("access_token", token);
  localStorage.setItem("user", JSON.stringify(user));
}

function clearAuth() {
  token = null;
  user = null;
  localStorage.removeItem("access_token");
  localStorage.removeItem("user");
  setAuthUI();
}

function isLoggedIn() {
  return token !== null && user !== null;
}

function requireAuth() {
  if (!isLoggedIn()) {
    openAuthPanel();
    return false;
  }
  return true;
}

function isAdmin() {
  return user && user.role === "admin";
}

function requireAdmin() {
  if (!isLoggedIn()) {
    openAuthPanel();
    return false;
  }
  if (!isAdmin()) {
    htmx.ajax("GET", "/pages/dashboard.html", {
      target: "#main-content",
      swap: "innerHTML",
    });
    return false;
  }
  return true;
}

function handleLogout() {
  if (token) {
    apiPost("/auth/logout");
  }
  clearAuth();
  htmx.ajax("GET", "/pages/home.html", {
    target: "#main-content",
    swap: "innerHTML",
  });
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Bookmarks  (API-backed)
// ═══════════════════════════════════════════════════════════════════════════════

// In-memory bookmark set for fast UI checks (synced from API)
var bookmarkIds = [];

function fetchBookmarks() {
  if (!token) {
    bookmarkIds = [];
    return Promise.resolve();
  }
  return apiGet("/users/me/bookmarks").then(function (res) {
    if (res.ok && res.body.success && res.body.data) {
      bookmarkIds = (res.body.data || []).map(function (b) {
        return b.circular_id;
      });
    } else {
      bookmarkIds = [];
    }
  });
}

function isBookmarked(id) {
  return bookmarkIds.indexOf(String(id)) > -1;
}

function toggleBookmark(id) {
  if (!requireAuth()) return;

  var wasBookmarked = isBookmarked(id);
  var method = wasBookmarked
    ? apiDelete("/users/me/bookmarks/" + id)
    : apiPost("/users/me/bookmarks/" + id);

  // Optimistic UI update
  if (wasBookmarked) {
    bookmarkIds = bookmarkIds.filter(function (bid) {
      return bid !== String(id);
    });
  } else {
    bookmarkIds.push(String(id));
  }

  method
    .then(function (res) {
      if (!res.ok) {
        // Revert on failure
        if (wasBookmarked) {
          bookmarkIds.push(String(id));
        } else {
          bookmarkIds = bookmarkIds.filter(function (bid) {
            return bid !== String(id);
          });
        }
      }
    })
    .catch(function () {
      // Revert on error
      if (wasBookmarked) {
        bookmarkIds.push(String(id));
      } else {
        bookmarkIds = bookmarkIds.filter(function (bid) {
          return bid !== String(id);
        });
      }
    });

  // Re-render affected UI
  if (document.getElementById("circular-list")) renderCirculars();
  if (
    document.getElementById("circular-detail-content") &&
    currentCircularId === id
  )
    fetchAndRenderDetail(id);
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Circulars  (API-backed)
// ═══════════════════════════════════════════════════════════════════════════════

var cachedCategories = [];

function fetchCategories() {
  return apiGet("/categories").then(function (res) {
    if (res.ok && res.body.success && res.body.data) {
      cachedCategories = res.body.data;
    } else {
      // Fallback: use demo data if available
      cachedCategories = typeof CATEGORIES !== "undefined" ? CATEGORIES : [];
    }
    return cachedCategories;
  });
}

var cachedFeatured = [];

function fetchFeatured() {
  return apiGet("/circulars/featured").then(function (res) {
    if (res.ok && res.body.success && res.body.data) {
      cachedFeatured = res.body.data;
    } else {
      cachedFeatured = [];
    }
    return cachedFeatured;
  });
}

function fetchCirculars(page, limit, filters) {
  var params = "?page=" + (page || 1) + "&limit=" + (limit || ITEMS_PER_PAGE);
  if (filters.search) params += "&search=" + encodeURIComponent(filters.search);
  if (filters.sort) params += "&sort=" + encodeURIComponent(filters.sort);
  if (filters.category)
    params += "&category=" + encodeURIComponent(filters.category);

  var container = document.getElementById("circular-list");
  if (container) {
    container.innerHTML =
      '<div class="loading-state"><span class="loading-spinner"></span><p>Loading circulars...</p></div>';
  }

  return apiGet("/circulars" + params).then(function (res) {
    if (res.ok && res.body.success && res.body.data) {
      return {
        items: res.body.data.items || [],
        pagination: res.body.data.pagination || {
          page: 1,
          limit: ITEMS_PER_PAGE,
          total: 0,
          total_pages: 0,
        },
      };
    }
    // Fallback to demo data
    if (typeof DEMO_CIRCULARS !== "undefined") {
      return filterDemoCirculars(page, limit, filters);
    }
    return {
      items: [],
      pagination: { page: 1, limit: ITEMS_PER_PAGE, total: 0, total_pages: 0 },
    };
  });
}

function fetchCircularDetail(id) {
  return apiGet("/circulars/" + id).then(function (res) {
    if (res.ok && res.body.success && res.body.data) {
      return res.body.data;
    }
    return null;
  });
}

// Fallback: filter demo data when API is unavailable
function filterDemoCirculars(page, limit, filters) {
  if (typeof DEMO_CIRCULARS === "undefined") {
    return {
      items: [],
      pagination: { page: 1, limit: limit, total: 0, total_pages: 0 },
    };
  }
  var filtered = DEMO_CIRCULARS.slice();
  var q = (filters.search || "").toLowerCase();
  if (q) {
    filtered = filtered.filter(function (c) {
      return (
        c.title.toLowerCase().indexOf(q) > -1 ||
        c.organization_name.toLowerCase().indexOf(q) > -1
      );
    });
  }
  if (filters.sort === "deadline_asc") {
    filtered.sort(function (a, b) {
      return (
        new Date(a.application_deadline) - new Date(b.application_deadline)
      );
    });
  } else {
    filtered.sort(function (a, b) {
      return new Date(b.published_date) - new Date(a.published_date);
    });
  }
  var total = filtered.length;
  var totalPages = Math.ceil(total / limit);
  var start = (page - 1) * limit;
  return {
    items: filtered.slice(start, start + limit),
    pagination: {
      page: page,
      limit: limit,
      total: total,
      total_pages: totalPages,
    },
  };
}

// ── Render helpers (API-backed) ────────────────────────────────────────────────

function renderHomeCategoryPills() {
  var c = document.getElementById("home-category-pills");
  if (!c) return;
  fetchCategories().then(function (cats) {
    c.innerHTML = cats
      .map(function (cat) {
        return (
          '<button class="cat-pill" data-slug="' +
          cat.slug +
          '">' +
          (cat.icon || "📋") +
          " " +
          cat.name +
          "</button>"
        );
      })
      .join("");
  });
}

function renderFeaturedCirculars() {
  var c = document.getElementById("featured-circulars");
  if (!c) return;
  fetchFeatured().then(function (items) {
    if (items.length === 0) {
      c.innerHTML =
        '<p style="padding:2rem;text-align:center;color:var(--ghost);">No featured circulars yet. Check back soon!</p>';
      return;
    }
    c.innerHTML = items.map(buildCard).join("");
  });
}

function renderCirculars() {
  var container = document.getElementById("circular-list");
  var empty = document.getElementById("empty-state");
  var pag = document.getElementById("pagination");
  if (!container) return;

  var filters = {
    search: currentFilters.search,
    sort: currentFilters.sort,
    category: currentFilters.category,
  };

  fetchCirculars(currentPage, ITEMS_PER_PAGE, filters).then(function (result) {
    var items = result.items;
    var pagination = result.pagination;

    if (empty) empty.style.display = items.length === 0 ? "block" : "none";
    container.innerHTML = items.map(buildCard).join("");
    renderPagination(pagination.total_pages);
  });
}

function renderCategoryPills() {
  var c = document.getElementById("category-pills");
  if (!c) return;
  fetchCategories().then(function (cats) {
    c.innerHTML = cats
      .map(function (cat) {
        return (
          '<button class="cat-pill" data-slug="' +
          cat.slug +
          '">' +
          (cat.icon || "📋") +
          " " +
          cat.name +
          "</button>"
        );
      })
      .join("");
  });
}

function renderPagination(totalPages) {
  var c = document.getElementById("pagination");
  if (!c || totalPages <= 1) {
    if (c) c.innerHTML = "";
    return;
  }
  var html = "";
  html +=
    '<button class="page-btn" data-page="' +
    (currentPage - 1) +
    '"' +
    (currentPage === 1 ? " disabled" : "") +
    ">Prev</button>";
  for (var i = 1; i <= totalPages; i++) {
    html +=
      '<button class="page-btn' +
      (i === currentPage ? " active" : "") +
      '" data-page="' +
      i +
      '">' +
      i +
      "</button>";
  }
  html +=
    '<button class="page-btn" data-page="' +
    (currentPage + 1) +
    '"' +
    (currentPage === totalPages ? " disabled" : "") +
    ">Next</button>";
  c.innerHTML = html;
}

// ── Card builder ───────────────────────────────────────────────────────────────
function buildCard(c) {
  var cat =
    c.category ||
    (c.category_id
      ? cachedCategories.find(function (x) {
          return x.id === c.category_id;
        })
      : null);
  var deadline = c.application_deadline;
  var days = deadline ? daysUntil(deadline) : 999;
  var bm = isBookmarked(c.id);
  var badgeClass = days <= 3 ? "now" : days <= 7 ? "soon" : "ok";
  var badgeText =
    days <= 0 ? "Expired" : days === 1 ? "1 day left" : days + " days left";
  return [
    '<div class="circular-card">',
    '<div class="card-top">',
    '<span class="category-badge">' +
      (cat ? (cat.icon || "") + " " + cat.name : c.organization_name) +
      "</span>",
    '<div style="display:flex;align-items:center;gap:0.35rem;">',
    '<span class="deadline-badge ' + badgeClass + '">' + badgeText + "</span>",
    '<button class="bookmark-btn" data-bookmark="' +
      c.id +
      '" aria-label="' +
      (bm ? "Remove" : "Add") +
      ' bookmark">' +
      (bm ? "❤️" : "🤍") +
      "</button>",
    "</div>",
    "</div>",
    '<a href="/circulars/' +
      c.id +
      '" class="card-title" data-circular-id="' +
      c.id +
      '">' +
      escapeHtml(c.title) +
      "</a>",
    '<div class="card-org">' + escapeHtml(c.organization_name) + "</div>",
    '<div class="card-meta"><span>' +
      (c.location || "Bangladesh") +
      "</span><span>" +
      (c.district || "") +
      "</span></div>",
    '<div class="card-footer"><span>' +
      (c.vacancy || "N/A") +
      " post" +
      (c.vacancy > 1 ? "s" : "") +
      "</span><span>" +
      formatDate(c.published_date) +
      "</span></div>",
    "</div>",
  ].join("");
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Circular Detail  (API-backed)
// ═══════════════════════════════════════════════════════════════════════════════

function renderCircularDetail() {
  fetchAndRenderDetail(currentCircularId);
}

function fetchAndRenderDetail(id) {
  var contentEl = document.getElementById("circular-detail-content");
  if (!contentEl) return;

  contentEl.innerHTML =
    '<div class="loading-state"><span class="loading-spinner"></span><p>Loading...</p></div>';

  fetchCircularDetail(id).then(function (c) {
    if (!c) {
      contentEl.innerHTML = "<p>Circular not found.</p>";
      return;
    }

    var cat =
      c.category ||
      (c.category_id
        ? cachedCategories.find(function (x) {
            return x.id === c.category_id;
          })
        : null);
    var deadline = c.application_deadline;
    var days = deadline ? daysUntil(deadline) : 999;
    var bm = isBookmarked(c.id);
    var badgeClass = days <= 3 ? "now" : days <= 7 ? "soon" : "ok";
    var badgeText =
      days <= 0 ? "Expired" : days === 1 ? "1 day left" : days + " days left";

    contentEl.innerHTML = [
      '<div class="detail-card">',
      '<div class="detail-header">',
      "<h2>" + escapeHtml(c.title) + "</h2>",
      '<div class="detail-badges">',
      '<span class="category-badge">' +
        (cat ? (cat.icon || "") + " " + cat.name : "") +
        "</span>",
      '<span class="deadline-badge ' +
        badgeClass +
        '">' +
        badgeText +
        "</span>",
      '<span class="status-badge ' +
        (c.status || "active") +
        '">' +
        (c.status || "active") +
        "</span>",
      "</div>",
      "</div>",
      '<div class="detail-org">' +
        escapeHtml(c.organization_name) +
        " · " +
        (c.location || "Bangladesh") +
        "</div>",
      '<div class="detail-grid">',
      '<div class="detail-item"><span class="label">Vacancy</span><span class="value">' +
        (c.vacancy || "N/A") +
        "</span></div>",
      '<div class="detail-item"><span class="label">Salary</span><span class="value">' +
        (c.salary_display || "Negotiable") +
        "</span></div>",
      '<div class="detail-item"><span class="label">Deadline</span><span class="value">' +
        (deadline ? formatDate(deadline) : "Not specified") +
        "</span></div>",
      '<div class="detail-item"><span class="label">Apply via</span><span class="value">' +
        (c.apply_via || "See details") +
        "</span></div>",
      c.education_level
        ? '<div class="detail-item"><span class="label">Education</span><span class="value">' +
          c.education_level +
          "</span></div>"
        : "",
      c.age_min
        ? '<div class="detail-item"><span class="label">Age range</span><span class="value">' +
          c.age_min +
          "–" +
          c.age_max +
          " years</span></div>"
        : "",
      c.job_type
        ? '<div class="detail-item"><span class="label">Job type</span><span class="value">' +
          c.job_type +
          "</span></div>"
        : "",
      c.gender
        ? '<div class="detail-item"><span class="label">Gender</span><span class="value">' +
          c.gender +
          "</span></div>"
        : "",
      "</div>",
      c.description
        ? '<div class="detail-description"><h4>Description</h4><p>' +
          escapeHtml(c.description) +
          "</p></div>"
        : "",
      c.requirements
        ? '<div class="detail-description"><h4>Requirements</h4><p>' +
          escapeHtml(c.requirements) +
          "</p></div>"
        : "",
      '<div class="detail-actions">',
      '<a href="' +
        (c.apply_url || "#") +
        '" target="_blank" rel="noopener" class="primary-btn">Apply now</a>',
      '<button class="secondary-btn" data-bookmark="' +
        c.id +
        '">' +
        (bm ? "❤️ Saved" : "🤍 Save this job") +
        "</button>",
      "</div>",
      "</div>",
    ].join("");
  });
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Dashboard  (API-backed)
// ═══════════════════════════════════════════════════════════════════════════════

function renderDashboard() {
  var nameEl = document.getElementById("dashboard-user-name");
  if (!requireAuth()) {
    if (nameEl) nameEl.textContent = "...";
    return;
  }
  if (nameEl) nameEl.textContent = user.name;

  // Show/hide admin panel
  var adminPanel = document.getElementById("admin-panel");
  if (adminPanel) {
    adminPanel.style.display = isAdmin() ? "block" : "none";
  }
  if (isAdmin()) {
    renderAdminDashboard();
  }

  fetchBookmarks().then(function () {
    var listEl = document.getElementById("bookmark-list");
    var noBmEl = document.getElementById("no-bookmarks");

    if (bookmarkIds.length === 0) {
      if (listEl) listEl.innerHTML = "";
      if (noBmEl) noBmEl.style.display = "block";
      return;
    }

    if (listEl) {
      listEl.innerHTML =
        '<div class="loading-state"><span class="loading-spinner"></span><p>Loading bookmarks...</p></div>';
    }

    var promises = bookmarkIds.map(function (id) {
      return fetchCircularDetail(id).catch(function () {
        return null;
      });
    });

    Promise.all(promises).then(function (circulars) {
      var valid = circulars.filter(function (c) {
        return c !== null;
      });
      if (listEl) {
        listEl.innerHTML = valid.map(buildCard).join("");
      }
      if (noBmEl) {
        noBmEl.style.display = valid.length === 0 ? "block" : "none";
      }
    });
  });
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Admin Dashboard
// ═══════════════════════════════════════════════════════════════════════════════

function fetchAdminStats() {
  return apiGet("/admin/stats").then(function (res) {
    if (res.ok && res.body.success) return res.body.data;
    return {};
  });
}

function fetchScrapeLogs() {
  return apiGet("/admin/scrape/logs").then(function (res) {
    if (res.ok && res.body.success) return res.body.data || [];
    return [];
  });
}

function triggerScrape() {
  var btn = document.getElementById("scrape-trigger-btn");
  var status = document.getElementById("scrape-status");
  if (btn) {
    btn.disabled = true;
    btn.textContent = "Running...";
  }
  if (status) status.textContent = "";

  apiPost("/admin/scrape/run")
    .then(function (res) {
      if (res.ok && res.body.success) {
        if (status) {
          status.textContent =
            res.body.data && res.body.data.message
              ? res.body.data.message
              : "Scrape triggered";
          status.style.color = "var(--success)";
        }
      } else {
        if (status) {
          status.textContent = "Failed to trigger scrape";
          status.style.color = "var(--error)";
        }
      }
      if (btn) {
        btn.disabled = false;
        btn.textContent = "Run Manual Scrape";
      }
    })
    .catch(function () {
      if (status) {
        status.textContent = "Network error";
        status.style.color = "var(--error)";
      }
      if (btn) {
        btn.disabled = false;
        btn.textContent = "Run Manual Scrape";
      }
    });
}

function renderAdminDashboard() {
  // Stats
  var statsEl = document.getElementById("admin-stats");
  if (statsEl) {
    fetchAdminStats().then(function (stats) {
      statsEl.innerHTML = [
        '<div class="stat-card">',
        '<span class="stat-value">' + (stats.total_circulars || 0) + "</span>",
        '<span class="stat-label">Total Circulars</span>',
        "</div>",
        '<div class="stat-card">',
        '<span class="stat-value">' + (stats.active_circulars || 0) + "</span>",
        '<span class="stat-label">Active</span>',
        "</div>",
        '<div class="stat-card">',
        '<span class="stat-value">' + (stats.total_users || 0) + "</span>",
        '<span class="stat-label">Users</span>',
        "</div>",
      ].join("");
    });
  }

  // Scrape logs
  var logsEl = document.getElementById("scrape-logs");
  if (logsEl) {
    fetchScrapeLogs().then(function (logs) {
      if (logs.length === 0) {
        logsEl.innerHTML =
          '<p style="color:var(--ghost);font-size:0.85rem;">No scrape runs yet.</p>';
        return;
      }
      logsEl.innerHTML = logs
        .slice(0, 10)
        .map(function (log) {
          var statusClass =
            log.status === "completed"
              ? "ok"
              : log.status === "failed"
                ? "now"
                : "soon";
          return [
            '<div class="scrape-log-row">',
            '<span class="deadline-badge ' +
              statusClass +
              '">' +
              log.status +
              "</span>",
            "<span>" + (log.source || "manual") + "</span>",
            "<span>Fetched: " + (log.total_fetched || 0) + "</span>",
            "<span>New: " + (log.new_inserted || 0) + "</span>",
            '<span style="color:var(--ghost);font-size:0.75rem;">' +
              formatDate(log.started_at) +
              "</span>",
            "</div>",
          ].join("");
        })
        .join("");
    });
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Admin Circular CRUD
// ═══════════════════════════════════════════════════════════════════════════════

function loadAdminCirculars() {
  var container = document.getElementById("admin-circular-list");
  if (!container) return;

  apiGet("/circulars?limit=100&status=all").then(function (res) {
    if (!res.ok || !res.body.success) {
      container.innerHTML = "<p>Failed to load circulars.</p>";
      return;
    }
    var items = res.body.data.items || [];
    if (items.length === 0) {
      container.innerHTML =
        "<p style='color:var(--ghost)'>No circulars found.</p>";
      return;
    }
    container.innerHTML =
      '<table class="admin-table"><thead><tr>' +
      "<th>Title</th><th>Org</th><th>Status</th><th>Featured</th><th>Actions</th>" +
      "</tr></thead><tbody>" +
      items
        .map(function (c) {
          return [
            "<tr>",
            "<td>" + escapeHtml(c.title) + "</td>",
            "<td>" + escapeHtml(c.organization_name) + "</td>",
            '<td><span class="status-badge ' +
              (c.status || "active") +
              '">' +
              (c.status || "active") +
              "</span></td>",
            "<td>" + (c.is_featured ? "⭐" : "—") + "</td>",
            '<td class="admin-actions">',
            '<button class="secondary-btn" onclick="editCircular(\'' +
              c.id +
              '\')" style="font-size:0.75rem;padding:0.25rem 0.5rem;">Edit</button>',
            '<button class="secondary-btn" onclick="toggleCircularFeature(\'' +
              c.id +
              '\')" style="font-size:0.75rem;padding:0.25rem 0.5rem;">' +
              (c.is_featured ? "Unfeature" : "Feature") +
              "</button>",
            '<button class="secondary-btn" onclick="deleteCircular(\'' +
              c.id +
              '\')" style="font-size:0.75rem;padding:0.25rem 0.5rem;color:var(--error);">Delete</button>',
            "</td>",
            "</tr>",
          ].join("");
        })
        .join("") +
      "</tbody></table>";
  });
}

function showCircularForm(id) {
  var panel = document.getElementById("circular-form-panel");
  if (!panel) return;
  panel.style.display = "block";

  // Reset form
  document.getElementById("circular-edit-id").value = "";
  document.getElementById("circ-title").value = "";
  document.getElementById("circ-org").value = "";
  document.getElementById("circ-cat").value = "";
  document.getElementById("circ-vacancy").value = "";
  document.getElementById("circ-location").value = "Bangladesh";
  document.getElementById("circ-pub").value = new Date()
    .toISOString()
    .split("T")[0];
  document.getElementById("circ-deadline").value = "";
  document.getElementById("circ-salary").value = "";
  document.getElementById("circ-apply-url").value = "";
  document.getElementById("circ-status").value = "active";
  document.getElementById("circ-featured").checked = false;
  document.getElementById("circ-desc").value = "";
  document.getElementById("circ-req").value = "";
  document.getElementById("circular-form-error").textContent = "";

  if (id) {
    document.getElementById("circular-form-title").textContent =
      "Edit Circular";
    document.getElementById("circular-save-btn").textContent = "Update";
    fetchCircularDetail(id).then(function (c) {
      if (!c) return;
      document.getElementById("circular-edit-id").value = c.id;
      document.getElementById("circ-title").value = c.title || "";
      document.getElementById("circ-org").value = c.organization_name || "";
      document.getElementById("circ-cat").value = c.category_id || "";
      document.getElementById("circ-vacancy").value = c.vacancy || "";
      document.getElementById("circ-location").value =
        c.location || "Bangladesh";
      if (c.published_date)
        document.getElementById("circ-pub").value =
          c.published_date.split("T")[0];
      if (c.application_deadline)
        document.getElementById("circ-deadline").value =
          c.application_deadline.split("T")[0];
      document.getElementById("circ-salary").value = c.salary_display || "";
      document.getElementById("circ-apply-url").value = c.apply_url || "";
      document.getElementById("circ-status").value = c.status || "active";
      document.getElementById("circ-featured").checked = !!c.is_featured;
      document.getElementById("circ-desc").value = c.description || "";
      document.getElementById("circ-req").value = c.requirements || "";
      panel.scrollIntoView({ behavior: "smooth" });
    });
  } else {
    document.getElementById("circular-form-title").textContent =
      "Create Circular";
    document.getElementById("circular-save-btn").textContent = "Save";
    panel.scrollIntoView({ behavior: "smooth" });
  }
}

function hideCircularForm() {
  var panel = document.getElementById("circular-form-panel");
  if (panel) panel.style.display = "none";
}

function saveCircular(evt) {
  evt.preventDefault();
  var errorEl = document.getElementById("circular-form-error");
  var id = document.getElementById("circular-edit-id").value;
  var body = {
    title: document.getElementById("circ-title").value.trim(),
    organization_name: document.getElementById("circ-org").value.trim(),
    category_id: parseInt(document.getElementById("circ-cat").value) || null,
    vacancy: parseInt(document.getElementById("circ-vacancy").value) || null,
    location:
      document.getElementById("circ-location").value.trim() || "Bangladesh",
    published_date: document.getElementById("circ-pub").value,
    application_deadline:
      document.getElementById("circ-deadline").value || null,
    salary_display: document.getElementById("circ-salary").value.trim() || null,
    apply_url: document.getElementById("circ-apply-url").value.trim() || null,
    status: document.getElementById("circ-status").value,
    is_featured: document.getElementById("circ-featured").checked,
    description: document.getElementById("circ-desc").value.trim() || null,
    requirements: document.getElementById("circ-req").value.trim() || null,
  };

  if (!body.title || !body.organization_name || !body.published_date) {
    if (errorEl)
      errorEl.textContent =
        "Title, Organization, and Published Date are required";
    return;
  }

  if (errorEl) errorEl.textContent = "Saving...";

  var method = id
    ? apiPut("/circulars/" + id, body)
    : apiPost("/circulars", body);

  method
    .then(function (res) {
      if (res.ok) {
        if (errorEl) errorEl.textContent = "";
        hideCircularForm();
        loadAdminCirculars();
      } else {
        if (errorEl)
          errorEl.textContent = (res.body && res.body.error) || "Save failed";
      }
    })
    .catch(function () {
      if (errorEl) errorEl.textContent = "Network error";
    });
}

function apiPut(path, body) {
  var headers = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = "Bearer " + token;
  return fetch(API_URL + path, {
    method: "PUT",
    headers: headers,
    body: JSON.stringify(body),
    credentials: "include",
  }).then(function (res) {
    return res.json().then(function (data) {
      return { ok: res.ok, status: res.status, body: data };
    });
  });
}

function editCircular(id) {
  showCircularForm(id);
}

function deleteCircular(id) {
  if (!confirm("Delete this circular? This cannot be undone.")) return;
  apiDelete("/circulars/" + id).then(function (res) {
    if (res.ok) {
      loadAdminCirculars();
    }
  });
}

function toggleCircularFeature(id) {
  apiPatch("/circulars/" + id + "/feature").then(function (res) {
    if (res.ok) {
      loadAdminCirculars();
    }
  });
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Filter / Search
// ═══════════════════════════════════════════════════════════════════════════════

function filterCirculars() {
  var si = document.getElementById("search-input");
  var ss = document.getElementById("sort-select");
  var catActive = document.querySelector("#category-pills .cat-pill.active");

  currentFilters.search = si ? si.value : window.currentSearchValue || "";
  currentFilters.sort = ss ? ss.value : "published_desc";
  currentFilters.category = catActive ? catActive.dataset.slug : "";
  currentPage = 1;
  renderCirculars();
}

// ═══════════════════════════════════════════════════════════════════════════════
//  Utilities
// ═══════════════════════════════════════════════════════════════════════════════

function formatDate(ds) {
  if (!ds) return "N/A";
  var d = new Date(ds);
  if (isNaN(d.getTime())) return "N/A";
  return d.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function daysUntil(ds) {
  if (!ds) return 999;
  var t = new Date(ds),
    n = new Date();
  if (isNaN(t.getTime())) return 999;
  return Math.ceil((t - n) / (1000 * 60 * 60 * 24));
}

function escapeHtml(str) {
  if (!str) return "";
  var d = document.createElement("div");
  d.textContent = str;
  return d.innerHTML;
}
