// ── State ──────────────────────────────────────────────────────────────────────
var API_URL = "http://localhost:8080/api/v1";
var ITEMS_PER_PAGE = 6;
var currentFilters = { search: "", sort: "published_desc" };
var currentPage = 1;
var currentCircularId = null;
window.currentSearchValue = null;

// ── Auth state ──────────────────────────────────────────────────────────────────
var token = localStorage.getItem("access_token") || null;
var user = JSON.parse(localStorage.getItem("user") || "null");

document.addEventListener('DOMContentLoaded', function() {
    setAuthUI();
    // Delegate click events for dynamic elements
    document.addEventListener('click', handleDynamicClick);
});

// ── Dynamic click delegation (bookmarks, pagination, cards, theme) ─────────────
function handleDynamicClick(evt) {
    // Theme toggle
    var themeBtn = evt.target.closest('#theme-toggle');
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
      swap: "outerHTML",
    });
    return;
  }
}

// ── HTMX auth interceptor ───────────────────────────────────────────────────────
document.body.addEventListener("htmx:responseError", function (evt) {
  if (evt.detail.xhr.status === 401) {
    fetch(API_URL + "/auth/refresh", { method: "POST", credentials: "include" })
      .then(function (res) {
        return res.json();
      })
      .then(function (data) {
        if (data.success) {
          token = data.data.access_token;
          user = data.data.user;
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

// Handle auth form responses
document.body.addEventListener("htmx:afterRequest", function (evt) {
  var formId = evt.detail.target && evt.detail.target.id;
  if (formId !== "login-form" && formId !== "register-form") return;
  var xhr = evt.detail.xhr;
  var errorEl = document.getElementById(
    formId === "login-form" ? "login-error" : "register-error",
  );
  try {
    var data = JSON.parse(xhr.responseText);
    if (xhr.status === 200 || xhr.status === 201) {
      if (data.success && data.data) {
        if (data.data.access_token) {
          token = data.data.access_token;
          user = data.data.user;
          saveAuth();
          setAuthUI();
          closeAuthPanel();
        } else if (data.data.id && formId === "register-form") {
          autoLoginAfterRegister();
        } else if (data.data.message && errorEl) {
          errorEl.textContent = data.data.message;
          errorEl.style.color = "var(--success)";
        }
      }
    } else if (data.error && errorEl) {
      errorEl.textContent = data.error;
      errorEl.style.color = "var(--error)";
    }
  } catch (e) {
    if (errorEl) {
      errorEl.textContent = "Something went wrong";
      errorEl.style.color = "var(--error)";
    }
  }
});

// ── Theme toggle ────────────────────────────────────────────────────────────────
function toggleTheme() {
    var html = document.documentElement;
    var current = html.getAttribute('data-theme') || 'light';
    var next = current === 'dark' ? 'light' : 'dark';
    html.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
    updateThemeIcons(next);
}

function updateThemeIcons(theme) {
    // Update all toggle buttons (navbar may have multiple after HTMX swaps)
    var btns = document.querySelectorAll('#theme-toggle');
    btns.forEach(function(btn) {
        btn.textContent = theme === 'dark' ? '🌙' : '☀️';
        btn.setAttribute('title', theme === 'dark' ? 'Switch to light' : 'Switch to dark');
        btn.setAttribute('aria-label', theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode');
    });
}

// Set initial icon state on any new theme-toggle that HTMX loads
document.body.addEventListener('htmx:afterSettle', function() {
    var theme = document.documentElement.getAttribute('data-theme') || 'light';
    updateThemeIcons(theme);
});
  btn.setAttribute(
    "title",
    theme === "dark" ? "Switch to light mode" : "Switch to dark mode",
  );
}
function autoLoginAfterRegister() {
  var email = document.getElementById("register-email")?.value;
  var pw = document.getElementById("register-password")?.value;
  if (!email || !pw) return;
  fetch(API_URL + "/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: email, password: pw }),
    credentials: "include",
  })
    .then(function (res) {
      return res.json();
    })
    .then(function (data) {
      if (data.success) {
        token = data.data.access_token;
        user = data.data.user;
        saveAuth();
        setAuthUI();
        closeAuthPanel();
      }
    })
    .catch(function () {});
}

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
    fetch(API_URL + "/auth/logout", {
      method: "POST",
      headers: { Authorization: "Bearer " + token },
    });
  }
  clearAuth();
  htmx.ajax("GET", "/pages/home.html", {
    target: "#main-content",
    swap: "outerHTML",
  });
}

// ── Bookmarks ────────────────────────────────────────────────────────────────────
function getBookmarks() {
  return JSON.parse(localStorage.getItem("bookmarks") || "[]");
}
function saveBookmarks(b) {
  localStorage.setItem("bookmarks", JSON.stringify(b));
}
function isBookmarked(id) {
  return getBookmarks().indexOf(id) > -1;
}

function toggleBookmark(id) {
  var bookmarks = getBookmarks();
  var idx = bookmarks.indexOf(id);
  if (idx > -1) bookmarks.splice(idx, 1);
  else bookmarks.push(id);
  saveBookmarks(bookmarks);
  if (document.getElementById("circular-list")) filterCirculars();
  if (
    document.getElementById("circular-detail-content") &&
    currentCircularId === id
  )
    renderCircularDetail();
}

// ── Home Category Pills ─────────────────────────────────────────────────────────
function renderHomeCategoryPills() {
  var c = document.getElementById("home-category-pills");
  if (!c) return;
  c.innerHTML = CATEGORY_FILTERS.map(function (cat) {
    return (
      '<button class="cat-pill" data-slug="' +
      cat.slug +
      '">' +
      cat.icon +
      " " +
      cat.name +
      "</button>"
    );
  }).join("");
}

// ── Featured ────────────────────────────────────────────────────────────────────
function renderFeaturedCirculars() {
  var c = document.getElementById("featured-circulars");
  if (!c) return;
  var featured = DEMO_CIRCULARS.filter(function (d) {
    return d.is_featured;
  }).slice(0, 5);
  c.innerHTML = featured.map(buildCard).join("");
}

// ── Circulars ───────────────────────────────────────────────────────────────────
function filterCirculars() {
  var si = document.getElementById("search-input");
  var ss = document.getElementById("sort-select");
  currentFilters.search = si ? si.value : (window.currentSearchValue || '');
  currentFilters.sort = ss ? ss.value : 'published_desc';
  currentPage = 1;
  renderCirculars();
}

function renderCirculars() {
  var container = document.getElementById("circular-list");
  var empty = document.getElementById("empty-state");
  var pag = document.getElementById("pagination");
  if (!container) return;

  // Fallback: ensure data is loaded
  if (typeof DEMO_CIRCULARS === 'undefined') {
    container.innerHTML = '<p style="padding:2rem;text-align:center;color:var(--ghost);">Loading circulars...</p>';
    return;
  }

  var filtered = DEMO_CIRCULARS.slice();
  var q = (currentFilters.search || "").toLowerCase();
  if (q) {
    filtered = filtered.filter(function (c) {
      return (
        c.title.toLowerCase().indexOf(q) > -1 ||
        c.organization_name.toLowerCase().indexOf(q) > -1
      );
    });
  }
  if (currentFilters.sort === "deadline_asc") {
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

  var totalPages = Math.ceil(filtered.length / ITEMS_PER_PAGE);
  if (currentPage > totalPages) currentPage = totalPages || 1;
  var start = (currentPage - 1) * ITEMS_PER_PAGE;
  var items = filtered.slice(start, start + ITEMS_PER_PAGE);

  if (empty) empty.style.display = filtered.length === 0 ? "block" : "none";
  container.innerHTML = items.map(buildCard).join("");
  renderPagination(totalPages);
}

// ── Card ────────────────────────────────────────────────────────────────────────
function buildCard(c) {
  var cat = CATEGORIES.find(function (x) {
    return x.id === c.category_id;
  });
  var days = daysUntil(c.application_deadline);
  var bm = isBookmarked(c.id);
  var badgeClass = days <= 3 ? "now" : days <= 7 ? "soon" : "ok";
  var badgeText =
    days <= 0 ? "Expired" : days === 1 ? "1 day left" : days + " days left";
  return [
    '<div class="circular-card">',
    '<div class="card-top">',
    '<span class="category-badge">' +
      (cat ? cat.icon : "") +
      " " +
      (cat ? cat.name : "") +
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
      c.location +
      "</span><span>" +
      (c.district || "") +
      "</span></div>",
    '<div class="card-footer"><span>' +
      c.vacancy +
      " post" +
      (c.vacancy > 1 ? "s" : "") +
      "</span><span>" +
      formatDate(c.published_date) +
      "</span></div>",
    "</div>",
  ].join("");
}

// ── Category Pills (circulars page) ─────────────────────────────────────────────
function renderCategoryPills() {
  var c = document.getElementById("category-pills");
  if (!c) return;
  c.innerHTML = CATEGORY_FILTERS.map(function (cat) {
    return (
      '<button class="cat-pill" data-slug="' +
      cat.slug +
      '">' +
      cat.icon +
      " " +
      cat.name +
      "</button>"
    );
  }).join("");
}

// ── Pagination ──────────────────────────────────────────────────────────────────
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

// ── Detail ──────────────────────────────────────────────────────────────────────
function renderCircularDetail() {
  var c = DEMO_CIRCULARS.find(function (x) {
    return x.id === currentCircularId;
  });
  if (!c) {
    document.getElementById("circular-detail-content").innerHTML =
      "<p>Circular not found.</p>";
    return;
  }
  var cat = CATEGORIES.find(function (x) {
    return x.id === c.category_id;
  });
  var days = daysUntil(c.application_deadline);
  var bm = isBookmarked(c.id);
  var badgeClass = days <= 3 ? "now" : days <= 7 ? "soon" : "ok";
  var badgeText =
    days <= 0 ? "Expired" : days === 1 ? "1 day left" : days + " days left";
  document.getElementById("circular-detail-content").innerHTML = [
    '<div class="detail-card">',
    '<div class="detail-header">',
    "<h2>" + escapeHtml(c.title) + "</h2>",
    '<div class="detail-badges">',
    '<span class="category-badge">' +
      (cat ? cat.icon : "") +
      " " +
      (cat ? cat.name : "") +
      "</span>",
    '<span class="deadline-badge ' + badgeClass + '">' + badgeText + "</span>",
    '<span class="status-badge ' + c.status + '">' + c.status + "</span>",
    "</div>",
    "</div>",
    '<div class="detail-org">' +
      escapeHtml(c.organization_name) +
      " · " +
      c.location +
      "</div>",
    '<div class="detail-grid">',
    '<div class="detail-item"><span class="label">Vacancy</span><span class="value">' +
      c.vacancy +
      "</span></div>",
    '<div class="detail-item"><span class="label">Salary</span><span class="value">' +
      (c.salary_display || "Negotiable") +
      "</span></div>",
    '<div class="detail-item"><span class="label">Deadline</span><span class="value">' +
      formatDate(c.application_deadline) +
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
      '" target="_blank" class="primary-btn">Apply now</a>',
    '<button class="secondary-btn" data-bookmark="' +
      c.id +
      '">' +
      (bm ? "❤️ Saved" : "🤍 Save this job") +
      "</button>",
    "</div>",
    "</div>",
  ].join("");
}

// ── Dashboard ───────────────────────────────────────────────────────────────────
function renderDashboard() {
  if (!requireAuth()) {
    document.getElementById("dashboard-user-name").textContent = "...";
    return;
  }
  document.getElementById("dashboard-user-name").textContent = user.name;
  var bookmarks = getBookmarks();
  if (bookmarks.length > 0) {
    var list = DEMO_CIRCULARS.filter(function (c) {
      return bookmarks.indexOf(c.id) > -1;
    });
    document.getElementById("bookmark-list").innerHTML = list
      .map(buildCard)
      .join("");
    document.getElementById("no-bookmarks").style.display = "none";
  } else {
    document.getElementById("bookmark-list").innerHTML = "";
    document.getElementById("no-bookmarks").style.display = "block";
  }
}

// ── Utilities ───────────────────────────────────────────────────────────────────
function formatDate(ds) {
  var d = new Date(ds);
  return d.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function daysUntil(ds) {
  var t = new Date(ds),
    n = new Date();
  return Math.ceil((t - n) / (1000 * 60 * 60 * 24));
}

function escapeHtml(str) {
  var d = document.createElement("div");
  d.textContent = str;
  return d.innerHTML;
}
