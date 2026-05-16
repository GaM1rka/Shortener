const shortenForm = document.querySelector("#shortenForm");
const analyticsForm = document.querySelector("#analyticsForm");
const urlInput = document.querySelector("#urlInput");
const aliasInput = document.querySelector("#aliasInput");
const analyticsCode = document.querySelector("#analyticsCode");
const result = document.querySelector("#result");
const shortUrl = document.querySelector("#shortUrl");
const copyButton = document.querySelector("#copyButton");
const apiStatus = document.querySelector("#apiStatus");
const totalClicks = document.querySelector("#totalClicks");
const originalUrl = document.querySelector("#originalUrl");
const byDay = document.querySelector("#byDay");
const byUserAgent = document.querySelector("#byUserAgent");
const clicks = document.querySelector("#clicks");
const toast = document.querySelector("#toast");

async function request(path, options = {}) {
  const response = await fetch(path, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    let message = `Request failed with ${response.status}`;
    try {
      const body = await response.json();
      message = body.error || message;
    } catch {
      // Keep the generic message.
    }
    throw new Error(message);
  }

  return response.json();
}

async function checkHealth() {
  try {
    await request("/healthz");
    apiStatus.textContent = "API online";
    apiStatus.className = "status ok";
  } catch (error) {
    apiStatus.textContent = "API unavailable";
    apiStatus.className = "status bad";
  }
}

shortenForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const payload = {
    url: urlInput.value.trim(),
  };

  const customAlias = aliasInput.value.trim();
  if (customAlias) {
    payload.custom_alias = customAlias;
  }

  try {
    const link = await request("/shorten", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    shortUrl.textContent = link.short_url;
    shortUrl.href = link.short_url;
    result.hidden = false;
    analyticsCode.value = link.short_code;
    await loadAnalytics(link.short_code);
    showToast("Short link created");
  } catch (error) {
    showToast(error.message, true);
  }
});

analyticsForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  await loadAnalytics(analyticsCode.value.trim());
});

copyButton.addEventListener("click", async () => {
  try {
    await navigator.clipboard.writeText(shortUrl.href);
    showToast("Copied");
  } catch (error) {
    showToast("Could not copy link", true);
  }
});

async function loadAnalytics(code) {
  if (!code) {
    return;
  }

  try {
    const analytics = await request(`/analytics/${encodeURIComponent(code)}`);
    totalClicks.textContent = analytics.total_clicks;
    originalUrl.textContent = analytics.original_url;
    originalUrl.href = analytics.original_url;
    renderBuckets(byDay, analytics.by_day);
    renderBuckets(byUserAgent, analytics.by_user_agent);
    renderClicks(analytics.clicks);
  } catch (error) {
    showToast(error.message, true);
  }
}

function renderBuckets(container, buckets) {
  container.innerHTML = "";

  if (!buckets || buckets.length === 0) {
    container.textContent = "No data";
    container.className = "bucket-list empty";
    return;
  }

  container.className = "bucket-list";
  for (const bucket of buckets) {
    const row = document.createElement("div");
    row.className = "bucket";
    row.innerHTML = `<span></span><strong></strong>`;
    row.querySelector("span").textContent = bucket.key;
    row.querySelector("strong").textContent = bucket.count;
    container.append(row);
  }
}

function renderClicks(items) {
  clicks.innerHTML = "";

  if (!items || items.length === 0) {
    clicks.textContent = "No data";
    clicks.className = "click-list empty";
    return;
  }

  clicks.className = "click-list";
  for (const click of items.slice(0, 20)) {
    const row = document.createElement("div");
    row.className = "click";
    row.innerHTML = `<span></span><strong></strong>`;
    row.querySelector("span").textContent = click.user_agent || "unknown";
    row.querySelector("strong").textContent = new Date(click.clicked_at).toLocaleString();
    clicks.append(row);
  }
}

function showToast(message, isError = false) {
  toast.textContent = message;
  toast.className = isError ? "toast error" : "toast";
  toast.hidden = false;
  window.clearTimeout(showToast.timeout);
  showToast.timeout = window.setTimeout(() => {
    toast.hidden = true;
  }, 3200);
}

checkHealth();
