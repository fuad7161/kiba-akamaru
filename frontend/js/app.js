// ═══════════════════════════════════════════════════════════════════════════════
//  JobCirculer — app.js  (API-backed)
// ═══════════════════════════════════════════════════════════════════════════════

// ── Configuration ──────────────────────────────────────────────────────────────
var API_URL = "http://localhost:8080/api/v1";
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

// ── Init ───────────────────────────────────────────────────────────────────────
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

// Auth form handlers (plain JS, not HTMX — full control over flow)

function handleLogin(evt) {
  evt.preventDefault();
  var email = document.getElementById("login-email")?.value?.trim();
  var password = document.getElementById("login-password")?.value;
  var errorEl = document.getElementById("login-error");

  if (!email || !password) {
    if (errorEl) {
      errorEl.textContent = "Email and password are required";
      errorEl.style.color = "var(--error)";
    }
    return;
  }

  if (errorEl) errorEl.textContent = "";

  apiPost("/auth/login", { email: email, password: password })
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
          errorEl.textContent = (res.body && res.body.error) || "Login failed";
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

function handleRegister(evt) {
  evt.preventDefault();
  var name = document.getElementById("register-name")?.value?.trim();
  var email = document.getElementById("register-email")?.value?.trim();
  var password = document.getElementById("register-password")?.value;
  var errorEl = document.getElementById("register-error");

  if (!name || !email || !password) {
    if (errorEl) {
      errorEl.textContent = "All fields are required";
      errorEl.style.color = "var(--error)";
    }
    return;
  }
  if (password.length < 6) {
    if (errorEl) {
      errorEl.textContent = "Password must be at least 6 characters";
      errorEl.style.color = "var(--error)";
    }
    return;
  }

  if (errorEl) errorEl.textContent = "";

  apiPost("/auth/register", { name: name, email: email, password: password })
    .then(function (res) {
      if (res.ok && res.body.success) {
        // Auto-login after successful registration
        apiPost("/auth/login", { email: email, password: password }).then(
          function (loginRes) {
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
          },
        );
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
  if (token && user) {
    if (gb) gb.style.display = "none";
    if (ud) ud.style.display = "flex";
    if (nd) nd.textContent = user.name;
    if (dl) dl.style.display = "inline";
  } else {
    if (gb) gb.style.display = "flex";
    if (ud) ud.style.display = "none";
    if (dl) dl.style.display = "none";
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

  fetchBookmarks().then(function () {
    var listEl = document.getElementById("bookmark-list");
    var noBmEl = document.getElementById("no-bookmarks");

    if (bookmarkIds.length === 0) {
      if (listEl) listEl.innerHTML = "";
      if (noBmEl) noBmEl.style.display = "block";
      return;
    }

    // For each bookmark, fetch the circular detail
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
